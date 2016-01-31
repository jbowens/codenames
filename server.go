package codenames

import (
	"encoding/json"
	"net/http"
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

	s.games = make(map[string]*Game)
	s.words = d.Words()
	s.Server.Handler = s.mux

	return s.Server.ListenAndServe()
}
