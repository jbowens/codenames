package codenames

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/jbowens/dictionary"
	"github.com/kr/pretty"
)

var gameIDs, words []string

func init() {
	dictGameIDs, err := dictionary.Load("assets/game-id-words.txt")
	if err != nil {
		panic(err)
	}
	dictWords, err := dictionary.Load("assets/original.txt")
	if err != nil {
		panic(err)
	}
	gameIDs = dictGameIDs.Words()
	words = dictWords.Words()
}

func randomGames(n int) map[string]*Game {
	games := make(map[string]*Game)
	for _, w := range gameIDs[:n] {
		games[w] = newGame(w, randomState(words), 0)
	}
	return games
}

func TestPersist(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-persist-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pebble store dir: %s\n", dir)

	var ps PebbleStore
	ps.DB, err = pebble.Open(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	games := randomGames(5)
	for _, g := range games {
		if err := ps.Save(g); err != nil {
			t.Fatal(err)
		}
	}
	if err := ps.DB.Close(); err != nil {
		t.Fatal(err)
	}

	// Re-open the DB.
	ps.DB, err = pebble.Open(dir, nil)
	if err != nil {
		t.Fatal(err)
	}

	restoredGames, err := ps.Restore()
	if err != nil {
		t.Fatal(err)
	}
	if err := ps.DB.Close(); err != nil {
		t.Fatal(err)
	}

	// Verify the game states are the same.
	for id, g := range games {
		got, ok := restoredGames[id]
		if !ok {
			t.Fatalf("restoredGames[%q] doesn't exist", id)
		}
		if !reflect.DeepEqual(got.GameState, g.GameState) {
			t.Fatalf("%s: GameStates don't match: %s, %s",
				id, pretty.Sprint(got.GameState), pretty.Sprint(g.GameState))
		}
		if !reflect.DeepEqual(got.Words, g.Words) {
			t.Fatalf("%s: Words don't match: %s, %s",
				id, pretty.Sprint(got.Words), pretty.Sprint(g.Words))
		}
		if !reflect.DeepEqual(got.Layout, g.Layout) {
			t.Fatalf("%s: Layout don't match: %s, %s",
				id, pretty.Sprint(got.Layout), pretty.Sprint(g.Layout))
		}

	}
}
