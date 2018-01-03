package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/jbowens/codenames"
	"github.com/jbowens/events"
)

func main() {
	eventLogger, err := events.Open("events.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}

	rand.Seed(time.Now().UnixNano())

	server := &codenames.Server{
		Server: http.Server{
			Addr: ":9091",
		},
		Events: eventLogger,
	}
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
}
