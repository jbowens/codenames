package main

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime/trace"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/jbowens/codenames"
	"github.com/pkg/errors"
)

const defaultListenAddr = ":9091"
const expiryDur = -24 * time.Hour

func main() {
	rand.Seed(time.Now().UnixNano())

	var bootstrapURL string
	var listenAddr string
	flag.StringVar(&listenAddr, "listen-addr", defaultListenAddr,
		"address for server to listen on")
	flag.StringVar(&bootstrapURL, "bootstrap-url", "",
		"URL of an existing codenames server to bootstrap the DB from")

	flag.Parse()

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

	if len(bootstrapURL) > 0 {
		err := bootstrap(bootstrapURL, dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Bootstrapping from %q: %s\n", bootstrapURL, err)
			os.Exit(1)
		}
		fmt.Printf("Bootstrapped from %q.\n", bootstrapURL)
		os.Exit(0)
	}

	var opts pebble.Options
	opts.EventListener = pebble.MakeLoggingEventListener(nil)
	opts.Experimental.DeleteRangeFlushDelay = 5 * time.Second
	opts.FormatMajorVersion = pebble.FormatMarkedCompacted
	opts.Levels = []pebble.LevelOptions{
		{BlockSize: 256 << 10, BlockRestartInterval: 256},
	}
	db, err := pebble.Open(dir, &opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pebble.Open: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ps := &codenames.PebbleStore{DB: db}

	// Delete any games created too long ago.
	err = ps.DeleteExpired(time.Now().Add(expiryDur))
	if err != nil {
		fmt.Fprintf(os.Stderr, "PebbleStore.DeletedExpired: %s\n", err)
		os.Exit(1)
	}
	go deleteExpiredPeriodically(ps)

	// Restore games from disk.
	games, err := ps.Restore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "PebbleStore.Restore: %s\n", err)
		os.Exit(1)
	}
	log.Printf("[STARTUP] Restored %d games from disk.\n", len(games))

	if traceDir := os.Getenv("TRACE"); len(traceDir) > 0 {
		log.Printf("[STARTUP] Traces enabled; storing most recent trace in %q", traceDir)
		go tracePeriodically(traceDir)
	}

	log.Printf("[STARTUP] Listening on addr %s\n", listenAddr)
	server := &codenames.Server{
		Server: http.Server{
			Addr: listenAddr,
		},
		Store: ps,
	}
	if err := server.Start(games); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
}

func bootstrap(bootstrapURL, dir string) error {
	ls, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(ls) > 0 {
		return fmt.Errorf("directory %q is not empty: aborting\n", dir)
	}
	u, err := url.Parse(bootstrapURL)
	if err != nil {
		return err
	}
	u.Path = "/checkpoint"
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("admin", os.Getenv("BOOTSTRAPPW"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("checkpoint returned %s status code\n", resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	r := bytes.NewReader(b)
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(gzr)
	for {
		var cf codenames.CheckpointFile
		err := dec.Decode(&cf)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(dir, cf.Name), cf.Data, os.ModePerm)
		if err != nil {
			return errors.Wrapf(err, "writing %s", filepath.Base(cf.Name))
		}
		log.Printf("Downloaded %s (%d bytes)\n", cf.Name, len(cf.Data))
	}
	return nil
}

func deleteExpiredPeriodically(ps *codenames.PebbleStore) {
	for range time.Tick(time.Hour) {
		err := ps.DeleteExpired(time.Now().Add(expiryDur))
		if err != nil {
			log.Printf("PebbleStore.DeletedExpired: %s\n", err)
		}
	}
}

func tracePeriodically(dst string) {
	for range time.Tick(time.Minute) {
		takeTrace(dst)
	}
}

func takeTrace(dst string) {
	f, err := ioutil.TempFile("", "trace")
	if err != nil {
		log.Printf("[TRACE] error creating temp file: %s", err)
		return
	}
	defer f.Close()

	err = trace.Start(f)
	if err != nil {
		log.Printf("[TRACE] error starting trace: %s", err)
		return
	}
	<-time.After(10 * time.Second)
	trace.Stop()
	err = os.Rename(f.Name(), dst)
	if err != nil {
		log.Printf("[TRACE] error renaming trace: %s", err)
	}
}
