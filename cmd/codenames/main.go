package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/jbowens/codenames"
)

const listenAddr = ":9091"
const expiryDur = -12 * time.Hour

func main() {
	rand.Seed(time.Now().UnixNano())

	// Open a Pebble DB to persist games to disk.
	dir := os.Getenv("PEBBLE_DIR")
	if dir == "" {
		dir = filepath.Join(".", "db")
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MkdirAll(%q): %s\n", dir, err)
		os.Exit(1)
	}
	log.Printf("[STARTUP] Opening pebble db from directory: %s\n", dir)

	db, err := pebble.Open(dir, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pebble.Open: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ps := &codenames.PebbleStore{DB: db}

	// Delete any games created over 24hrs ago.
	err = ps.DeleteExpired(time.Now().Add(expiryDur))
	if err != nil {
		fmt.Fprintf(os.Stderr, "PebbleStore.DeletedExpired: %s\n", err)
		os.Exit(1)
	}
	go deleteExpiredPeriodically(ps)

	// Restore games from disk.
	games, err := ps.Restore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "PebbleStore.Resore: %s\n", err)
		os.Exit(1)
	}
	log.Printf("[STARTUP] Restored %d games from disk.\n", len(games))

	log.Printf("[STARTUP] Listening on addr %s\n", listenAddr)
	server := &codenames.Server{
		Server: http.Server{Addr: listenAddr},
		Store:  ps,
	}
	if err := server.Start(games); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
}

func deleteExpiredPeriodically(ps *codenames.PebbleStore) {
	for range time.Tick(time.Hour) {
		err := ps.DeleteExpired(time.Now().Add(expiryDur))
		if err != nil {
			log.Printf("PebbleStore.DeletedExpired: %s\n", err)
		}
	}
}
