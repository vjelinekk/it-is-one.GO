package main

import (
	"log"

	"github.com/vjelinekk/it-is-one.GO/pkg/db"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"github.com/vjelinekk/it-is-one.GO/pkg/server"
)

func main() {
	// Initialize Database
	database, err := db.Init("data.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run Migrations
	log.Println("Running database migrations...")
	if err := database.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	srv := server.New("0.0.0.0:8080", database)

	log.Printf("Starting server on port 8080...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
