package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

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
		return "neutral"
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
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	StartingTeam Team      `json:"starting_team"`
	WinningTeam  *Team     `json:"winning_team,omitempty"`
	Round        int       `json:"round"`
	Clues        []Clue    `json:"clues"`
	Words        []string  `json:"words"`
	Layout       []Team    `json:"layout"`
	Revealed     []bool    `json:"revealed"`
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

func (g *Game) NextTurn() error {
	if g.WinningTeam != nil {
		return errors.New("the game is already over")
	}
	g.Round++
	return nil
}

func (g *Game) Guess(idx int) error {
	if idx > len(g.Layout) || idx < 0 {
		return fmt.Errorf("Index %d is invalid.", idx)
	}
	if g.Revealed[idx] {
		return errors.New("That cell has already been revealed.")
	}
	g.Revealed[idx] = true

	if g.Layout[idx] == Black {
		winners := g.CurrentTeam().Other()
		g.WinningTeam = &winners
		return nil
	}

	var remaining bool
	for i, t := range g.Layout {
		if t == g.CurrentTeam() && !g.Revealed[i] {
			remaining = true
		}
	}
	if !remaining {
		winners := g.CurrentTeam()
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

func newGame(name string, words []string) *Game {
	game := &Game{
		ID:           uuid.NewV4().String(),
		Name:         name,
		CreatedAt:    time.Now(),
		StartingTeam: Team(rand.Intn(2)) + Red,
		Words:        make([]string, 0, wordsPerGame),
		Layout:       make([]Team, 0, wordsPerGame),
		Revealed:     make([]bool, wordsPerGame),
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

	shuffleCount := rand.Intn(5) + 5
	for i := 0; i < shuffleCount; i++ {
		shuffle(teamAssignments)
	}
	game.Layout = teamAssignments

	return game
}

func shuffle(teamAssignments []Team) {
	for i := range teamAssignments {
		j := rand.Intn(i + 1)
		teamAssignments[i], teamAssignments[j] = teamAssignments[j], teamAssignments[i]
	}
}
