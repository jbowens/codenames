package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jbowens/codenames"
)

func main() {
	server := &codenames.Server{
		Server: http.Server{
			Addr: ":9090",
		},
	}
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}
}
