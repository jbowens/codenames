package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"

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
	Black
)

func (t Team) String() string {
	switch t {
	case Red:
		return "red"
	case Blue:
		return "blue"
	case Black:
		return "black"
	default:
		return "netural"
	}
}

func (t Team) Other() Team {
	if t == Red {
		return Blue
	}
	if t == Blue {
		return Red
	}
	return t
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

type Clue struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func (c Clue) String() string {
	return fmt.Sprintf("%s, %s", c.Word, c.Count)
}

type Game struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	StartingTeam Team               `json:"starting_team"`
	WinningTeam  *Team              `json:"winning_team,omitempty"`
	Round        int                `json:"round"`
	Clues        []Clue             `json:"clues"`
	Words        []string           `json:"words"`
	Layout       []Team             `json:"layout"`
	Revealed     []bool             `json:"revealed"`
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

	if codemaster && p.Team != Neutral && p.Team != Black {
		g.Codemaster[p.Team] = p
	}
	return nil
}

func (g *Game) ProvideClue(c Clue) error {
	if len(g.Clues) > g.Round {
		return fmt.Errorf("The clue %s was already provided this round.", g.Clues[g.Round])
	}
	if len(strings.Split(c.Word, " ")) > 1 {
		return errors.New("You must provide a single word as your clue.")
	}

	g.Clues = append(g.Clues, c)
	return nil
}

func (g *Game) Guess(idx int) error {
	if g.Revealed[idx] {
		return errors.New("That cell has already been revealed.")
	}
	g.Revealed[idx] = true

	if g.Layout[idx] == Black {
		winners := g.CurrentTeam().Other()
		g.WinningTeam = &winners
		return nil
	}

	if g.Layout[idx] != g.CurrentTeam() {
		g.Round = g.Round + 1
	}

	return nil
}

func (g *Game) CurrentTeam() Team {
	if g.Round%2 == 0 {
		return g.StartingTeam
	}
	return g.StartingTeam.Other()
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
		Revealed:     make([]bool, 0, wordsPerGame),
		Codemaster:   make(map[Team]*Player),
		Players:      make(map[string]*Player),
	}

	for i := range game.Revealed {
		game.Revealed[i] = false
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
	teamAssignments = append(teamAssignments, Red.Repeat(8)...)
	teamAssignments = append(teamAssignments, Blue.Repeat(8)...)
	teamAssignments = append(teamAssignments, Neutral.Repeat(7)...)
	teamAssignments = append(teamAssignments, Black)
	teamAssignments = append(teamAssignments, game.StartingTeam)
	for i := range teamAssignments {
		j := rand.Intn(i + 1)
		teamAssignments[i], teamAssignments[j] = teamAssignments[j], teamAssignments[i]
	}
	game.Layout = teamAssignments

	return game
}
