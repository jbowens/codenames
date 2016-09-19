package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
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

type Game struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	StartingTeam Team      `json:"starting_team"`
	WinningTeam  *Team     `json:"winning_team,omitempty"`
	Round        int       `json:"round"`
	Words        []string  `json:"words"`
	Layout       []Team    `json:"layout"`
	Revealed     []bool    `json:"revealed"`
	Clue         *Clue     `json:"clue"`
}

func (g *Game) checkWinningCondition() {
	if g.WinningTeam != nil {
		return
	}
	var redRemaining, blueRemaining bool
	for i, t := range g.Layout {
		if g.Revealed[i] {
			continue
		}
		if t == Red {
			redRemaining = true
		}
		if t == Blue {
			blueRemaining = true
		}
	}
	if !redRemaining {
		winners := Red
		g.WinningTeam = &winners
	}
	if !blueRemaining {
		winners := Blue
		g.WinningTeam = &winners
	}
}

func (g *Game) NextTurn() error {
	if g.WinningTeam != nil {
		return errors.New("game is already over")
	}
	g.Round++
	g.Clue = nil
	return nil
}

func (g *Game) Guess(idx int) error {
	if idx > len(g.Layout) || idx < 0 {
		return fmt.Errorf("index %d is invalid", idx)
	}
	if g.Revealed[idx] {
		return errors.New("cell has already been revealed")
	}
	g.Revealed[idx] = true

	if g.Layout[idx] == Black {
		winners := g.CurrentTeam().Other()
		g.WinningTeam = &winners
		return nil
	}

	g.checkWinningCondition()
	if g.Layout[idx] != g.CurrentTeam() {
		g.Round = g.Round + 1
		g.Clue = nil
	}
	return nil
}

func (g *Game) CurrentTeam() Team {
	if g.Round%2 == 0 {
		return g.StartingTeam
	}
	return g.StartingTeam.Other()
}

func (g *Game) AddClue(word string, count int) {
	g.Clue = &Clue{
		Word:  word,
		Count: count,
	}
}

func newGame(id string, words []string) *Game {
	game := &Game{
		ID:           id,
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
