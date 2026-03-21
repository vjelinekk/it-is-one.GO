package models

import "time"

type Caregiver struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	PatientID uint   `json:"patient_id"` // ref: > users.id
	Email     string `json:"email"`
}
