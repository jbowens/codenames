package codenames

import (
	"encoding/json"
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
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	StartingTeam Team     `json:"starting_team"`
	Round        int      `json:"round"`
	Words        []string `json:"words"`
	Layout       []Team   `json:"layout"`
}

func newGame(name string, words []string) *Game {
	game := &Game{
		ID:           uuid.NewV4().String(),
		Name:         name,
		StartingTeam: Team(rand.Intn(2)) + Red,
		Words:        make([]string, 0, wordsPerGame),
		Layout:       make([]Team, 0, wordsPerGame),
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
