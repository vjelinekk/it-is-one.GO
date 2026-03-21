package models

type Schedule struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	UserID uint   `json:"user_id"`
	Time1  string `json:"time1" gorm:"type:time"` // e.g. '08:00:00'
	Time2  string `json:"time2" gorm:"type:time"` // e.g. '20:00:00'
}
