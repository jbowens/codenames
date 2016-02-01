package codenames

import (
	"encoding/json"
	"net/http"
	"path"
	"sync"

	"github.com/jbowens/dictionary"
)

type Server struct {
	Server http.Server

	mu    sync.Mutex
	games map[string]*Game
	words []string
	mux   *http.ServeMux
}

// POST /new
func (s *Server) handleNewGame(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var request struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	g := newGame(request.Name, s.words)
	s.games[g.ID] = g

	writeJSON(rw, g)
}

// POST /join
func (s *Server) handleJoinGame(rw http.ResponseWriter, req *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var request struct {
		GameID     string  `json:"game_id"`
		Codemaster bool    `json:"codemaster"`
		Player     *Player `json:"player"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	g, ok := s.games[request.GameID]
	if !ok {
		http.Error(rw, "No such game", 404)
		return
	}

	err := g.AddPlayer(request.Player, request.Codemaster)
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	writeJSON(rw, g)
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

func (s *Server) Start() error {
	d, err := dictionary.Default()
	if err != nil {
		return err
	}
	d = dictionary.Filter(d, func(w string) bool {
		return len(w) < 12
	})

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/new", s.handleNewGame)
	s.mux.HandleFunc("/join", s.handleJoinGame)
	s.mux.HandleFunc("/game/", s.handleRetrieveGame)

	s.games = make(map[string]*Game)
	s.words = d.Words()
	s.Server.Handler = s.mux

	return s.Server.ListenAndServe()
}
