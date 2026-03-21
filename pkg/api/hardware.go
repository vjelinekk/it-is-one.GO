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
// @Accept json
// @Param body body object true "Battery Level"
// @Success 200 {string} string "OK"
// @Router /device/heartbeat [post]
func (h *HardwareHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	serial := r.Context().Value(DeviceSerialKey).(string)

	var req struct {
		BatteryLevel int `json:"battery_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	now := time.Now()
	// Update battery level and last seen for the user owning this device
	res := h.DB.Model(&models.User{}).Where("device_serial = ?", serial).Updates(map[string]interface{}{
		"device_battery":   req.BatteryLevel,
		"device_last_seen": &now,
	})

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
