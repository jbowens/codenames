package codenames

import (
	"encoding/json"
	"errors"
	"math/rand"

	"github.com/satori/go.uuid"
)

const (
	wordsPerGame = 25
)

type Team int

const (
	Neutral Team = iota
	Red
	Blue
)

func (t Team) String() string {
	switch t {
	case Red:
		return "red"
	case Blue:
		return "blue"
	default:
		return "netural"
	}
}

func (t Team) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t Team) Repeat(n int) []Team {
	s := make([]Team, n)
	for i := 0; i < n; i++ {
		s[i] = t
	}
	return s
}

type Game struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	StartingTeam Team               `json:"starting_team"`
	Round        int                `json:"round"`
	Words        []string           `json:"words"`
	Layout       []Team             `json:"layout"`
	Players      map[string]*Player `json:"players"`
	Codemaster   map[Team]*Player   `json:"codemaster"`
}

func (g *Game) AddPlayer(p *Player, codemaster bool) error {
	if g.Round > 0 {
		return errors.New("That game has already started.")
	}

	if p.ID == "" {
		p.ID = uuid.NewV4().String()
	}

	g.Players[p.ID] = p

	if codemaster && p.Team != Neutral {
		g.Codemaster[p.Team] = p
	}
	return nil
}

func (g *Game) CurrentTeam() Team {
	if g.Round%2 == 0 {
		return g.StartingTeam
	}
	if g.StartingTeam == Red {
		return Blue
	}
	return Red
}

type Player struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Team Team   `json:"team"`
}

func newGame(name string, words []string) *Game {
	game := &Game{
		ID:           uuid.NewV4().String(),
		Name:         name,
		StartingTeam: Team(rand.Intn(2)) + Red,
		Words:        make([]string, 0, wordsPerGame),
		Layout:       make([]Team, 0, wordsPerGame),
		Codemaster:   make(map[Team]*Player),
		Players:      make(map[string]*Player),
	}

	// Pick 25 random words.
	used := map[string]struct{}{}
	for len(used) < wordsPerGame {
		w := words[rand.Intn(len(words))]
		if _, ok := used[w]; !ok {
			used[w] = struct{}{}
			game.Words = append(game.Words, w)
		}
	}

	// Pick a random permutation of team assignments.
	var teamAssignments []Team
	teamAssignments = append(teamAssignments, Neutral.Repeat(8)...)
	teamAssignments = append(teamAssignments, Red.Repeat(8)...)
	teamAssignments = append(teamAssignments, Blue.Repeat(8)...)
	teamAssignments = append(teamAssignments, game.StartingTeam)
	for i := range teamAssignments {
		j := rand.Intn(i + 1)
		teamAssignments[i], teamAssignments[j] = teamAssignments[j], teamAssignments[i]
	}
	game.Layout = teamAssignments

	return game
}
