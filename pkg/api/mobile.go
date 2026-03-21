package api

import "gorm.io/gorm"

type MobileHandler struct {
	DB *gorm.DB
}

func NewMobileHandler(db *gorm.DB) *MobileHandler {
	return &MobileHandler{DB: db}
}
