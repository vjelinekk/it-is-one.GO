package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	FullName string `json:"full_name"`
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Timezone string `json:"timezone" gorm:"help:Necessary to know when it is 8 AM for THIS user"`

	// Hardware fields
	DeviceSerial   *string    `gorm:"uniqueIndex" json:"device_serial"`
	DeviceBattery  int        `gorm:"default:100" json:"device_battery"`
	DeviceLastSeen *time.Time `json:"device_last_seen"`

	// Notification thresholds
	NotifyAfterMinutes          int `gorm:"default:10" json:"notify_after_minutes"`
	NotifyCaregiversAfterRetries int `gorm:"default:3" json:"notify_caregivers_after_retries"`
	CurrentMissedDoses          int `gorm:"default:0" json:"current_missed_doses"`

	// Relationships
	Schedules  []Schedule  `json:"schedules,omitempty"`
	Caregivers []Caregiver `json:"caregivers,omitempty" gorm:"foreignKey:PatientID"`
}
