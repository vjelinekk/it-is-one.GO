package main

import (
	"log"

	"github.com/vjelinekk/it-is-one.GO/pkg/server"
)

func main() {
	srv := server.New(":8080")

	log.Printf("Starting server on port 8080...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
