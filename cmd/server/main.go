package main

import (
	"log"

	"github.com/vjelinekk/it-is-one.GO/pkg/db"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"github.com/vjelinekk/it-is-one.GO/pkg/server"
)

// @title Smart Pill Doser API
// @version 1.0
// @description API for Smart Pill Doser POC.
// @BasePath /

// @securityDefinitions.apikey MobileAuth
// @in header
// @name X-User-ID

// @securityDefinitions.apikey HardwareAuth
// @in header
// @name X-Device-Serial

func main() {
	// Initialize Database
	database, err := db.Init("data.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run Migrations
	log.Println("Running database migrations...")
	if err := database.AutoMigrate(
		&models.User{},
		&models.Schedule{},
		&models.IntakeLog{},
		&models.Caregiver{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	srv := server.New("0.0.0.0:8080", database)

	// Start Background Workers
	server.StartEscalator(database)
	server.StartWatchdog(database)

	log.Printf("Starting server on port 8080...")
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
