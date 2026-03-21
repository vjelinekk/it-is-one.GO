package models

type Schedule struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        uint   `json:"user_id"`
	ScheduledTime string `json:"scheduled_time" gorm:"type:time"` // e.g. '08:00:00'
}
