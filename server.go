package codenames

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/jbowens/assets"
	"github.com/jbowens/dictionary"
)

type Server struct {
	Server http.Server

	tpl   *template.Template
	jslib assets.Bundle
	js    assets.Bundle
	css   assets.Bundle

	gameIDWords []string

	mu    sync.Mutex
	games map[string]*Game
	words []string
	mux   *http.ServeMux
}

// GET /game/<id>
func (s *Server) handleRetrieveGame(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gameID := path.Base(req.URL.Path)
	g, ok := s.games[gameID]
	if ok {
		writeJSON(rw, g)
		return
	}

	g = newGame(gameID, s.words, rand.Int63())
	s.games[gameID] = g
	writeJSON(rw, g)
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

	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.games[request.GameID]
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	if err := g.Guess(request.Index); err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeJSON(rw, g)
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

	s.mu.Lock()
	defer s.mu.Unlock()

	g, ok := s.games[request.GameID]
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	if err := g.NextTurn(); err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeJSON(rw, g)
}

func (s *Server) handleNextGame(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID string `json:"game_id"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	g := newGame(request.GameID, s.words, rand.Int63())
	s.games[request.GameID] = g
	writeJSON(rw, g)
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
	s.jslib, err = assets.Development("assets/jslib")
	if err != nil {
		return err
	}
	s.js, err = assets.Development("assets/javascript")
	if err != nil {
		return err
	}
	s.css, err = assets.Development("assets/stylesheets")
	if err != nil {
		return err
	}

	s.mux = http.NewServeMux()

	s.mux.HandleFunc("/stats", s.handleStats)
	s.mux.HandleFunc("/next-game", s.handleNextGame)
	s.mux.HandleFunc("/end-turn", s.handleEndTurn)
	s.mux.HandleFunc("/guess", s.handleGuess)
	s.mux.HandleFunc("/game/", s.handleRetrieveGame)

	s.mux.Handle("/js/lib/", http.StripPrefix("/js/lib/", s.jslib))
	s.mux.Handle("/js/", http.StripPrefix("/js/", s.js))
	s.mux.Handle("/css/", http.StripPrefix("/css/", s.css))
	s.mux.HandleFunc("/", s.handleIndex)

	gameIDs = dictionary.Filter(gameIDs, func(s string) bool { return len(s) > 3 })
	s.gameIDWords = gameIDs.Words()

	s.games = make(map[string]*Game)
	s.words = d.Words()
	s.Server.Handler = s.mux

	go func() {
		for range time.Tick(10 * time.Minute) {
			s.cleanupOldGames()
		}
	}()
	return s.Server.ListenAndServe()
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
