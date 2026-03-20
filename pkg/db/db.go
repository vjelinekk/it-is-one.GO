package db

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Init initializes the SQLite database
func Init(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to database at: %s", dbPath)
	return db, nil
}
