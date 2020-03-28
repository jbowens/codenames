package codenames

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jbowens/dictionary"
)

type Server struct {
	Server http.Server

	tpl         *template.Template
	gameIDWords []string

	mu           sync.Mutex
	games        map[string]*Game
	defaultWords []string
	mux          *http.ServeMux
}

func (s *Server) getGame(gameID, stateID string) (*Game, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getGameLocked(gameID, stateID)
}

func (s *Server) getGameLocked(gameID, stateID string) (*Game, bool) {
	g, ok := s.games[gameID]
	if ok {
		return g, ok
	}
	state, ok := decodeGameState(stateID, s.defaultWords)
	if !ok {
		return nil, false
	}
	g = newGame(gameID, state)
	s.games[gameID] = g
	return g, true
}

// GET /game/<id>
// (deprecated: use POST /game-state instead)
func (s *Server) handleRetrieveGame(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := req.ParseForm()
	if err != nil {
		http.Error(rw, "Error decoding query string", 400)
		return
	}

	gameID := path.Base(req.URL.Path)
	g, ok := s.getGameLocked(gameID, req.Form.Get("state_id"))
	if ok {
		writeGame(rw, g)
		return
	}

	g = newGame(gameID, randomState(s.defaultWords))
	s.games[gameID] = g
	writeGame(rw, g)
}

// POST /game-state
func (s *Server) handleGameState(rw http.ResponseWriter, req *http.Request) {
	var body struct {
		GameID  string `json:"game_id"`
		StateID string `json:"state_id"`
	}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(rw, "Error decoding request body", 400)
		return
	}

	s.mu.Lock()
	g, ok := s.getGameLocked(body.GameID, body.StateID)
	if ok {
		s.mu.Unlock()
		writeGame(rw, g)
		return
	}

	g = newGame(body.GameID, randomState(s.defaultWords))
	s.games[body.GameID] = g
	s.mu.Unlock()
	writeGame(rw, g)
}

// POST /guess
func (s *Server) handleGuess(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID  string `json:"game_id"`
		StateID string `json:"state_id"`
		Index   int    `json:"index"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	g, ok := s.getGame(request.GameID, request.StateID)
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	if err := g.Guess(request.Index); err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeGame(rw, g)
}

// POST /end-turn
func (s *Server) handleEndTurn(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID  string `json:"game_id"`
		StateID string `json:"state_id"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	g, ok := s.getGame(request.GameID, request.StateID)
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	if err := g.NextTurn(); err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeGame(rw, g)
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

	g, ok := s.games[request.GameID]
	if !ok || request.CreateNew {
		g = newGame(request.GameID, randomState(words))
		s.games[request.GameID] = g
	}
	writeGame(rw, g)
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
	for _, g := range s.games {
		g.mu.Lock()
		if g.WinningTeam == nil && g.anyRevealed() {
			inProgress++
		}
		if hourAgo.Before(g.CreatedAt) {
			createdWithinAnHour++
		}
		g.mu.Unlock()
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
	for id, g := range s.games {
		g.mu.Lock()
		if g.WinningTeam != nil && g.CreatedAt.Add(3*time.Hour).Before(time.Now()) {
			delete(s.games, id)
			fmt.Printf("Removed completed game %s\n", id)
		} else if g.CreatedAt.Add(12 * time.Hour).Before(time.Now()) {
			delete(s.games, id)
			fmt.Printf("Removed expired game %s\n", id)
		}
		g.mu.Unlock()
	}
}

func (s *Server) Start() error {
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
	s.mux.HandleFunc("/game/", s.handleRetrieveGame)
	s.mux.HandleFunc("/game-state", s.handleGameState)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/dist"))))
	s.mux.HandleFunc("/", s.handleIndex)

	gameIDs = dictionary.Filter(gameIDs, func(s string) bool { return len(s) > 3 })
	s.gameIDWords = gameIDs.Words()

	s.games = make(map[string]*Game)
	s.defaultWords = d.Words()
	sort.Strings(s.defaultWords)
	s.Server.Handler = withPProfHandler(s.mux)

	go func() {
		for range time.Tick(10 * time.Minute) {
			s.cleanupOldGames()
		}
	}()

	fmt.Println("Started server. Available on http://localhost:9091")
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

func writeGame(rw http.ResponseWriter, g *Game) {
	writeJSON(rw, struct {
		*Game
		StateID string `json:"state_id"`
	}{g, g.GameState.ID()})
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
