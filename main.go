// Shop API — a small gin service used to show what Routy can do.
//
//	go run .            # listens on :8080
//	routy run api       # every request and flow against it
//	routy check         # .routy files, environments, formatting and Go routes
package main

import (
	"log"
	"os"

	"github.com/1rowvy/routy-example-gin/internal/api"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("shop api on %s", addr)
	if err := api.Router().Run(addr); err != nil {
		log.Fatal(err)
	}
}
