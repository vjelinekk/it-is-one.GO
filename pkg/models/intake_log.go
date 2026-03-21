package models

import "time"

type IntakeLog struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserID   uint      `json:"user_id"`
	TakenAt  time.Time `json:"taken_at"`
}
