package codenames

import (
	"encoding/json"
	"html/template"
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

	mu    sync.Mutex
	games map[string]*Game
	words []string
	mux   *http.ServeMux
}

// POST /new
func (s *Server) handleNewGame(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	g := newGame(request.Name, s.words)
	s.games[g.ID] = g

	writeJSON(rw, g)
}

// GET /games
func (s *Server) handleListGames(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	games := make([]*Game, 0, len(s.games))
	for _, g := range s.games {
		if g.WinningTeam == nil && g.CreatedAt.Add(5*time.Minute).After(time.Now()) {
			games = append(games, g)
		}
	}
	writeJSON(rw, games)
}

// GET /game/<id>
func (s *Server) handleRetrieveGame(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gameID := path.Base(req.URL.Path)
	g, ok := s.games[gameID]
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

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

// POST /clue
func (s *Server) handleClue(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID string `json:"game_id"`
		Clue   Clue   `json:"clue"`
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

	if err := g.ProvideClue(request.Clue); err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeJSON(rw, g)
}

func (s *Server) Start() error {
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

	s.mux.HandleFunc("/games", s.handleListGames)
	s.mux.HandleFunc("/new", s.handleNewGame)
	s.mux.HandleFunc("/end-turn", s.handleEndTurn)
	s.mux.HandleFunc("/guess", s.handleGuess)
	s.mux.HandleFunc("/game/", s.handleRetrieveGame)

	s.mux.Handle("/js/lib/", http.StripPrefix("/js/lib/", s.jslib))
	s.mux.Handle("/js/", http.StripPrefix("/js/", s.js))
	s.mux.Handle("/css/", http.StripPrefix("/css/", s.css))
	s.mux.HandleFunc("/", s.handleIndex)

	s.games = make(map[string]*Game)
	s.words = d.Words()
	s.Server.Handler = s.mux

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
