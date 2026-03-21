package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"gorm.io/gorm"
)

type HardwareHandler struct {
	DB *gorm.DB
}

func NewHardwareHandler(db *gorm.DB) *HardwareHandler {
	return &HardwareHandler{DB: db}
}

// Heartbeat updates the device status
// @Summary Device heartbeat
// @Tags Hardware
// @Security HardwareAuth
// @Security MobileAuth
// @Accept json
// @Param body body object true "Battery Level"
// @Success 200 {string} string "OK"
// @Router /api/v1/device/heartbeat [post]
func (h *HardwareHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BatteryLevel int `json:"battery_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"device_battery":   req.BatteryLevel,
		"device_last_seen": &now,
	}

	var res *gorm.DB
	if serial, ok := r.Context().Value(DeviceSerialKey).(string); ok && serial != "" {
		res = h.DB.Model(&models.User{}).Where("device_serial = ?", serial).Updates(updates)
	} else {
		userID := r.Context().Value(UserIDKey).(uint)
		res = h.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates)
	}

	if res.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if res.RowsAffected == 0 {
		http.Error(w, "Device not linked to any user", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}
