package codenames

import (
	"encoding/json"
	"testing"

	"github.com/jbowens/dictionary"
)

func BenchmarkGameMarshal(b *testing.B) {
	b.StopTimer()
	d, err := dictionary.Load("assets/original.txt")
	if err != nil {
		b.Fatal(err)
	}
	g := newGame("foo", GameState{
		Seed:     1,
		Round:    0,
		Revealed: make([]bool, 25),
		WordSet:  d.Words(),
	}, GameOptions{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err = json.Marshal(g)
		if err != nil {
			b.Fatal(err)
		}
	}
}
