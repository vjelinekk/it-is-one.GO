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

// LogIntake records a medication intake event
// @Summary Log intake
// @Tags Hardware
// @Security HardwareAuth
// @Accept json
// @Param body body object true "Intake details"
// @Success 201 {string} string "Created"
// @Router /device/intake [post]
func (h *HardwareHandler) LogIntake(w http.ResponseWriter, r *http.Request) {
	serial := r.Context().Value(DeviceSerialKey).(string)

	var req struct {
		Timestamp time.Time `json:"timestamp"`
		Status    string    `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Find user associated with the device
	var user models.User
	if err := h.DB.Where("device_serial = ?", serial).First(&user).Error; err != nil {
		http.Error(w, "Device not linked to any user", http.StatusNotFound)
		return
	}

	// Create the intake log. In a full system, we would match this against the ScheduleID
	// For this POC, we record the user, status, and actual timestamp.
	logEntry := models.IntakeLog{
		UserID:    user.ID,
		Status:    req.Status,
		ActualAt:  &req.Timestamp,
		PlannedAt: req.Timestamp, // POC simplified mapping
	}

	if err := h.DB.Create(&logEntry).Error; err != nil {
		http.Error(w, "Failed to log intake", http.StatusInternalServerError)
		return
	}

	// Reset missed doses count on success
	h.DB.Model(&user).Update("current_missed_doses", 0)

	w.WriteHeader(http.StatusCreated)
}
