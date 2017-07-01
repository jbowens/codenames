package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jbowens/codenames"
	"github.com/jbowens/events"
)

func main() {
	eventLogger, err := events.Open("events.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}

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
