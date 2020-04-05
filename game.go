package codenames

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const wordsPerGame = 25

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

func (t *Team) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	switch s {
	case "red":
		*t = Red
	case "blue":
		*t = Blue
	case "black":
		*t = Black
	default:
		*t = Neutral
	}
	return nil
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

// GameState encapsulates enough data to reconstruct
// a Game's state. It's used to recreate games after
// a process restart.
type GameState struct {
	Seed      int64    `json:"seed"`
	PermIndex int      `json:"perm_index"`
	Round     int      `json:"round"`
	Revealed  []bool   `json:"revealed"`
	WordSet   []string `json:"word_set"`
}

func (gs GameState) anyRevealed() bool {
	var revealed bool
	for _, r := range gs.Revealed {
		revealed = revealed || r
	}
	return revealed
}

func randomState(words []string) GameState {
	return GameState{
		Seed:      rand.Int63(),
		PermIndex: 0,
		Revealed:  make([]bool, wordsPerGame),
		WordSet:   words,
	}
}

// nextGameState returns a new GameState for the next game.
func nextGameState(state GameState) GameState {
	state.PermIndex = state.PermIndex + wordsPerGame
	if state.PermIndex+wordsPerGame >= len(state.WordSet) {
		state.Seed = rand.Int63()
		state.PermIndex = 0
	}
	state.Revealed = make([]bool, wordsPerGame)
	return state
}

type Game struct {
	GameState
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	StartingTeam Team      `json:"starting_team"`
	WinningTeam  *Team     `json:"winning_team,omitempty"`
	Words        []string  `json:"words"`
	Layout       []Team    `json:"layout"`
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
		switch t {
		case Red:
			redRemaining = true
		case Blue:
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
		winners := g.currentTeam().Other()
		g.WinningTeam = &winners
		return nil
	}

	g.checkWinningCondition()
	if g.Layout[idx] != g.currentTeam() {
		g.Round = g.Round + 1
	}
	return nil
}

func (g *Game) currentTeam() Team {
	if g.Round%2 == 0 {
		return g.StartingTeam
	}
	return g.StartingTeam.Other()
}

func newGame(id string, state GameState) *Game {
	// consistent randomness across games with the same seed
	seedRnd := rand.New(rand.NewSource(state.Seed))
	// distinct randomness across games with same seed
	randRnd := rand.New(rand.NewSource(state.Seed * int64(state.PermIndex+1)))

	game := &Game{
		ID:           id,
		CreatedAt:    time.Now(),
		StartingTeam: Team(randRnd.Intn(2)) + Red,
		Words:        make([]string, 0, wordsPerGame),
		Layout:       make([]Team, 0, wordsPerGame),
		GameState:    state,
	}

	// Pick the next `wordsPerGame` words from the
	// randomly generated permutation
	perm := seedRnd.Perm(len(state.WordSet))
	permIndex := state.PermIndex
	for _, i := range perm[permIndex : permIndex+wordsPerGame] {
		w := state.WordSet[perm[i]]
		game.Words = append(game.Words, w)
	}

	// Pick a random permutation of team assignments.
	var teamAssignments []Team
	teamAssignments = append(teamAssignments, Red.Repeat(8)...)
	teamAssignments = append(teamAssignments, Blue.Repeat(8)...)
	teamAssignments = append(teamAssignments, Neutral.Repeat(7)...)
	teamAssignments = append(teamAssignments, Black)
	teamAssignments = append(teamAssignments, game.StartingTeam)

	shuffleCount := randRnd.Intn(5) + 5
	for i := 0; i < shuffleCount; i++ {
		shuffle(randRnd, teamAssignments)
	}
	game.Layout = teamAssignments
	return game
}

func shuffle(rnd *rand.Rand, teamAssignments []Team) {
	for i := range teamAssignments {
		j := rnd.Intn(i + 1)
		teamAssignments[i], teamAssignments[j] = teamAssignments[j], teamAssignments[i]
	}
}
