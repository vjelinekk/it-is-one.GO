package models

import "time"

type IntakeLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	ScheduleID uint      `json:"schedule_id"`
	UserID     uint      `json:"user_id"`
	Status     string    `json:"status"` // 'taken' or 'missed'
	PlannedAt  time.Time `json:"planned_at"`
	ActualAt   *time.Time `json:"actual_at" gorm:"help:When the ESP32 actually detected the pill was taken"`
}
