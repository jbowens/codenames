package codenames

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/jbowens/dictionary"
)

type Server struct {
	Server http.Server
	tpl    *template.Template

	gameIDWords []string

	mu           sync.Mutex
	games        map[string]*Game
	defaultWords []string
	mux          *http.ServeMux
}

func (s *Server) getGame(gameID, stateID string) (*Game, bool) {
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
	g, ok := s.getGame(gameID, req.Form.Get("state_id"))
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
	defer s.mu.Unlock()
	g, ok := s.getGame(body.GameID, body.StateID)
	if ok {
		writeGame(rw, g)
		return
	}
	g = newGame(body.GameID, randomState(s.defaultWords))
	s.games[body.GameID] = g
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

	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.mu.Lock()
	defer s.mu.Unlock()

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
		GameID  string   `json:"game_id"`
		WordSet []string `json:"word_set"`
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}
	if len(request.WordSet) > 0 && len(request.WordSet) < 25 {
		http.Error(rw, "Need at least 25 words", 400)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	words := request.WordSet
	sort.Strings(words)
	if len(words) == 0 {
		words = s.defaultWords
	}

	g := newGame(request.GameID, randomState(words))
	s.games[request.GameID] = g
	writeGame(rw, g)
}

type statsResponse struct {
	InProgress int `json:"games_in_progress"`
}

func (s *Server) handleStats(rw http.ResponseWriter, req *http.Request) {
	var inProgress int

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, g := range s.games {
		if g.WinningTeam == nil {
			inProgress++
		}
	}
	writeJSON(rw, statsResponse{inProgress})
}

func (s *Server) cleanupOldGames() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, g := range s.games {
		if g.WinningTeam != nil && g.CreatedAt.Add(12*time.Hour).Before(time.Now()) {
			delete(s.games, id)
			fmt.Printf("Removed completed game %s\n", id)
			continue
		}
		if g.CreatedAt.Add(24 * time.Hour).Before(time.Now()) {
			delete(s.games, id)
			fmt.Printf("Removed expired game %s\n", id)
			continue
		}
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
	s.Server.Handler = s.mux

	go func() {
		for range time.Tick(10 * time.Minute) {
			s.cleanupOldGames()
		}
	}()
	return s.Server.ListenAndServe()
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
