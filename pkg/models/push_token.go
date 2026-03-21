package models

import "time"

type PushToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID uint   `json:"user_id"`
	Token  string `gorm:"uniqueIndex;not null" json:"token"`
	Platform string `json:"platform"` // ios, android
}
