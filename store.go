package codenames

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/cockroachdb/pebble"
)

// PebbleStore wraps a *pebble.DB with an implementation of the
// Store interface, persisting games under a []byte(`/games/`)
// key prefix.
type PebbleStore struct {
	DB *pebble.DB
}

// Restore loads all persisted games from storage.
func (ps *PebbleStore) Restore() (map[string]*Game, error) {
	iter := ps.DB.NewIter(&pebble.IterOptions{
		LowerBound: []byte("/games/"),
		UpperBound: []byte(fmt.Sprintf("/games/%019d", math.MaxInt64)),
	})
	defer iter.Close()

	games := make(map[string]*Game)
	for _ = iter.First(); iter.Valid(); iter.Next() {
		var g Game
		err := json.Unmarshal(iter.Value(), &g)
		if err != nil {
			return nil, fmt.Errorf("Unmarshal game: %w", err)
		}
		games[g.ID] = &g
	}
	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("restore iter: %w", err)
	}
	return games, nil
}

// DeleteExpired deletes all games created before `expiry.`
func (ps *PebbleStore) DeleteExpired(expiry time.Time) error {
	return ps.DB.DeleteRange(
		mkkey(0, ""),
		mkkey(expiry.Unix(), ""),
		nil,
	)
}

// Save saves the game to persistent storage.
func (ps *PebbleStore) Save(g *Game) error {
	k, v, err := gameKV(g)
	if err != nil {
		return fmt.Errorf("trySave: %w", err)
	}

	err = ps.DB.Set(k, v, &pebble.WriteOptions{Sync: true})
	if err != nil {
		return fmt.Errorf("db.Set: %w", err)
	}
	return err
}

func gameKV(g *Game) (key, value []byte, err error) {
	value, err = json.Marshal(g)
	if err != nil {
		return nil, nil, fmt.Errorf("marshaling GameState: %w", err)
	}
	return mkkey(g.CreatedAt.Unix(), g.ID), value, nil
}

func mkkey(unixSecs int64, id string) []byte {
	// We could use a binary encoding for keys,
	// but it's not like we're storing that many
	// kv pairs. Ease of debugging is probably
	// more important.
	return []byte(fmt.Sprintf("/games/%019d/%q", unixSecs, id))
}

type discardStore struct{}

func (ds discardStore) Save(*Game) error { return nil }
