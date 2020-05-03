package codenames

import (
	"bytes"
	"encoding/binary"
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

func (ps *PebbleStore) CounterAdd(stat string, v int64) error {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutVarint(b[:], v)

	k := fmt.Sprintf("/stats/counters/%s", stat)
	return ps.DB.Merge([]byte(k), b[:n], nil)
}

func (ps *PebbleStore) GetCounter(statPrefix string) (int64, error) {
	prefix := []byte(fmt.Sprintf("/stats/counters/%s", statPrefix))

	iter := ps.DB.NewIter(nil)
	iter.SeekGE(prefix)

	var sum int64
	for ; iter.Valid() && bytes.HasPrefix(iter.Key(), prefix); iter.Next() {
		rawV := iter.Value()
		v, n := binary.Varint(rawV)
		if n < 0 {
			return 0, fmt.Errorf("unable to read stat value: %v for key %q", rawV, iter.Key())
		}
		sum += v
	}
	err := iter.Error()
	if closeErr := iter.Close(); closeErr != nil {
		err = closeErr
	}

	return sum, err
}

// PebbleMerge implements the pebble.Merge function type.
func PebbleMerge(k, v []byte) (pebble.ValueMerger, error) {
	vInt, n := binary.Varint(v)
	if n < 0 {
		return nil, fmt.Errorf("unable to read merge value: %v", v)
	}
	//if bytes.HasPrefix(k, []byte("/stats/counters/")) {
	return &addValueMerger{v: vInt}, nil
	//}
	//return nil, fmt.Errorf("unrecognized merge key: %s", pretty.Sprint(k))
}

// addValueMerger implements pebble.ValueMerger by interpreting values as a
// signed varint and adding its operands.
type addValueMerger struct {
	v int64
}

func (m *addValueMerger) MergeNewer(value []byte) error {
	v, n := binary.Varint(value)
	if n < 0 {
		return fmt.Errorf("unable to read merge value: %v", value)
	}
	m.v = m.v + v
	return nil
}

func (m *addValueMerger) MergeOlder(value []byte) error {
	v, n := binary.Varint(value)
	if n < 0 {
		return fmt.Errorf("unable to read merge value: %v", value)
	}
	m.v = m.v + v
	return nil
}

func (m *addValueMerger) Finish() ([]byte, error) {
	b := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(b, m.v)
	return b[:n], nil
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

func (_ discardStore) Save(*Game) error                 { return nil }
func (_ discardStore) GetCounter(string) (int64, error) { return 0, nil }
func (_ discardStore) CounterAdd(string, int64) error   { return nil }
