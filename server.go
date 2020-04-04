package codenames

import (
	"crypto/subtle"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jbowens/dictionary"
)

type Server struct {
	Server http.Server
	Store  Store

	tpl         *template.Template
	gameIDWords []string

	mu           sync.Mutex
	games        map[string]*GameHandle
	defaultWords []string
	mux          *http.ServeMux
}

type Store interface {
	Save(*Game) error
}

type GameHandle struct {
	store Store

	mu        sync.Mutex
	marshaled []byte
	g         *Game
}

func newHandle(g *Game, s Store) *GameHandle {
	gh := &GameHandle{store: s, g: g}
	err := s.Save(g)
	if err != nil {
		log.Printf("Unable to write updated game %q to disk: %s\n", gh.g.ID, err)
	}
	return gh
}

func (gh *GameHandle) update(fn func(*Game)) {
	gh.mu.Lock()
	defer gh.mu.Unlock()
	fn(gh.g)
	gh.marshaled = nil

	// write the updated game to disk
	err := gh.store.Save(gh.g)
	if err != nil {
		log.Printf("Unable to write updated game %q to disk: %s\n", gh.g.ID, err)
	}
}

// MarshalJSON implements the encoding/json.Marshaler interface.
// It caches a marshalled value of the game object.
func (gh *GameHandle) MarshalJSON() ([]byte, error) {
	gh.mu.Lock()
	defer gh.mu.Unlock()

	var err error
	if gh.marshaled == nil {
		gh.marshaled, err = json.Marshal(gh.g)
	}
	return gh.marshaled, err
}

func (s *Server) getGame(gameID string) (*GameHandle, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getGameLocked(gameID)
}

func (s *Server) getGameLocked(gameID string) (*GameHandle, bool) {
	gh, ok := s.games[gameID]
	if ok {
		return gh, ok
	}
	gh = newHandle(newGame(gameID, randomState(s.defaultWords)), s.Store)
	s.games[gameID] = gh
	return gh, true
}

// POST /game-state
func (s *Server) handleGameState(rw http.ResponseWriter, req *http.Request) {
	var body struct {
		GameID string `json:"game_id"`
	}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(rw, "Error decoding request body", 400)
		return
	}

	s.mu.Lock()
	gh, ok := s.getGameLocked(body.GameID)
	if ok {
		s.mu.Unlock()
		writeGame(rw, gh)
		return
	}

	gh = newHandle(newGame(body.GameID, randomState(s.defaultWords)), s.Store)
	s.games[body.GameID] = gh
	s.mu.Unlock()
	writeGame(rw, gh)
}

// POST /guess
func (s *Server) handleGuess(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID string `json:"game_id"`
		Index  int    `json:"index"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	gh, ok := s.getGame(request.GameID)
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	var err error
	gh.update(func(g *Game) {
		err = g.Guess(request.Index)
	})
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeGame(rw, gh)
}

// POST /end-turn
func (s *Server) handleEndTurn(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID string `json:"game_id"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	gh, ok := s.getGame(request.GameID)
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	var err error
	gh.update(func(g *Game) {
		err = g.NextTurn()
	})
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeGame(rw, gh)
}

func (s *Server) handleNextGame(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID    string   `json:"game_id"`
		WordSet   []string `json:"word_set"`
		CreateNew bool     `json:"create_new"`
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}
	wordSet := map[string]bool{}
	for _, w := range request.WordSet {
		wordSet[strings.TrimSpace(strings.ToUpper(w))] = true
	}
	if len(wordSet) > 0 && len(wordSet) < 25 {
		http.Error(rw, "Need at least 25 words", 400)
		return
	}

	var gh *GameHandle
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		words := s.defaultWords
		if len(wordSet) > 0 {
			words = nil
			for w := range wordSet {
				words = append(words, w)
			}
			sort.Strings(words)
		}

		gh, ok := s.games[request.GameID]
		if !ok {
			// no game exists, create for the first time
			gh = newHandle(newGame(request.GameID, randomState(words)), s.Store)
			s.games[request.GameID] = gh
		} else if request.CreateNew {
			nextState := nextGameState(gh.g.GameState)
			gh = newHandle(newGame(request.GameID, nextState), s.Store)
			s.games[request.GameID] = gh
		}
	}()
	writeGame(rw, gh)
}

type statsResponse struct {
	GamesTotal          int `json:"games_total"`
	GamesInProgress     int `json:"games_in_progress"`
	GamesCreatedOneHour int `json:"games_created_1h"`
}

func (s *Server) handleStats(rw http.ResponseWriter, req *http.Request) {
	hourAgo := time.Now().Add(-time.Hour)

	s.mu.Lock()
	defer s.mu.Unlock()

	var inProgress, createdWithinAnHour int
	for _, gh := range s.games {
		gh.mu.Lock()
		if gh.g.WinningTeam == nil && gh.g.anyRevealed() {
			inProgress++
		}
		if hourAgo.Before(gh.g.CreatedAt) {
			createdWithinAnHour++
		}
		gh.mu.Unlock()
	}
	writeJSON(rw, statsResponse{
		GamesTotal:          len(s.games),
		GamesInProgress:     inProgress,
		GamesCreatedOneHour: createdWithinAnHour,
	})
}

func (s *Server) cleanupOldGames() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, gh := range s.games {
		gh.mu.Lock()
		if gh.g.WinningTeam != nil && gh.g.CreatedAt.Add(3*time.Hour).Before(time.Now()) {
			delete(s.games, id)
			log.Printf("Removed completed game %s\n", id)
		} else if gh.g.CreatedAt.Add(12 * time.Hour).Before(time.Now()) {
			delete(s.games, id)
			log.Printf("Removed expired game %s\n", id)
		}
		gh.mu.Unlock()
	}
}

func (s *Server) Start(games map[string]*Game) error {
	gameIDs, err := dictionary.Load("assets/game-id-words.txt")
	if err != nil {
		return err
	}
	d, err := dictionary.Load("assets/original.txt")
	if err != nil {
		return err
	}
	s.tpl, err = template.New("index").Parse(tpl)
	if err != nil {
		return err
	}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/stats", s.handleStats)
	s.mux.HandleFunc("/next-game", s.handleNextGame)
	s.mux.HandleFunc("/end-turn", s.handleEndTurn)
	s.mux.HandleFunc("/guess", s.handleGuess)
	s.mux.HandleFunc("/game-state", s.handleGameState)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/dist"))))
	s.mux.HandleFunc("/", s.handleIndex)

	gameIDs = dictionary.Filter(gameIDs, func(s string) bool { return len(s) > 3 })
	s.gameIDWords = gameIDs.Words()

	s.games = make(map[string]*GameHandle)
	s.defaultWords = d.Words()
	sort.Strings(s.defaultWords)
	s.Server.Handler = withPProfHandler(s.mux)

	if s.Store == nil {
		s.Store = discardStore{}
	}

	if games != nil {
		for _, g := range games {
			s.games[g.ID] = newHandle(g, s.Store)
		}
	}

	go func() {
		for range time.Tick(10 * time.Minute) {
			s.cleanupOldGames()
		}
	}()

	return s.Server.ListenAndServe()
}

func withPProfHandler(next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	pprofHandler := basicAuth(mux, os.Getenv("PPROFPW"), "admin")

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/debug/pprof") {
			pprofHandler.ServeHTTP(rw, req)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func basicAuth(handler http.Handler, password, realm string) http.Handler {
	p := []byte(password)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(pass), p) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			io.WriteString(w, "Unauthorized\n")
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func writeGame(rw http.ResponseWriter, gh *GameHandle) {
	writeJSON(rw, gh)
}

func writeJSON(rw http.ResponseWriter, resp interface{}) {
	j, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, "unable to marshal response: "+err.Error(), 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(j)
}
