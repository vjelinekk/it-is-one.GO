package models

import "time"

type IntakeLog struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	UserID   uint      `json:"user_id"`
	DoseSlot int       `json:"dose_slot"` // 1 = first dose of day, 2 = second dose
	Date     string    `json:"date"`       // YYYY-MM-DD in user's timezone
	TakenAt  time.Time `json:"taken_at"`
}
