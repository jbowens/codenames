package codenames

import (
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
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

// Delete removes a game from persistent storage.
func (ps *PebbleStore) Delete(g *Game) error {
	k := mkkey(g.CreatedAt.Unix(), g.ID)
	err := ps.DB.Delete(k, nil)
	if err != nil {
		return fmt.Errorf("db.Delete: %w", err)
	}
	return nil
}

type CheckpointFile struct {
	Name string
	Data []byte
}

// Checkpoint returns a serialized represenation of the entire store.
func (ps *PebbleStore) Checkpoint(w io.Writer) error {
	// Compact the entire key space. The database tends to be small and there
	// tends to be a significant number of obsolete keys, so this shouldn't be
	// too expensive but will reduce the number of bytes we need to send over
	// the network.
	err := ps.DB.Compact([]byte{}, []byte{0xFF, 0xFF, 0xFF, 0xFF})
	if err != nil {
		return err
	}

	// Create a Pebble checkpoint in a temporary directory.
	name, err := ioutil.TempDir("", "checkpoint")
	if err != nil {
		return err
	}
	if err := os.RemoveAll(name); err != nil {
		return err
	}
	defer os.RemoveAll(name)

	err = ps.DB.Checkpoint(name)
	if err != nil {
		return err
	}

	// Write all the files in the checkpoint out over the network.
	gzipWriter := gzip.NewWriter(w)
	enc := gob.NewEncoder(gzipWriter)
	err = filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(name, path)
		if err != nil {
			return err
		}
		log.Printf("Checkpoint sending file %s (%d bytes)\n", relPath, len(b))
		return enc.Encode(CheckpointFile{
			Name: relPath,
			Data: b,
		})
	})
	if err != nil {
		return err
	}
	return gzipWriter.Close()
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

func (ds discardStore) Save(*Game) error           { return nil }
func (ds discardStore) Delete(*Game) error         { return nil }
func (ds discardStore) Checkpoint(io.Writer) error { return nil }
