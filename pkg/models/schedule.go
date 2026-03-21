package models

import "time"

type Schedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID        uint   `json:"user_id"`
	ScheduledTime string `json:"scheduled_time" gorm:"type:time"` // e.g. '08:00:00'
	DaysOfWeek    string `json:"days_of_week" gorm:"help:e.g. 1,2,3,4,5,6,7"`

	// Relationships
	IntakeLogs []IntakeLog `json:"intake_logs,omitempty"`
}
