package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user record in the database
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Name      string         `json:"name"`
}
