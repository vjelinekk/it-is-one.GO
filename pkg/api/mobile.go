package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"gorm.io/gorm"
)

type MobileHandler struct {
	DB *gorm.DB
}

func NewMobileHandler(db *gorm.DB) *MobileHandler {
	return &MobileHandler{DB: db}
}

// User & Device Linking
// GetMe gets the current user's profile
// @Summary Get user profile
// @Tags Mobile
// @Security MobileAuth
// @Produce json
// @Success 200 {object} models.User
// @Router /users/me [get]
func (h *MobileHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var user models.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// UpdateMe updates the current user's profile
// @Summary Update user profile
// @Tags Mobile
// @Security MobileAuth
// @Accept json
// @Param body body object true "Update payload"
// @Success 200 {string} string "OK"
// @Router /users/me [put]
func (h *MobileHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req struct {
		FullName            *string `json:"full_name"`
		Timezone            *string `json:"timezone"`
		MissedDoseThreshold *int    `json:"missed_dose_threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.MissedDoseThreshold != nil {
		updates["missed_dose_threshold"] = *req.MissedDoseThreshold
	}

	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// LinkDevice links an ESP32 device to the user
// @Summary Link device
// @Tags Mobile
// @Security MobileAuth
// @Accept json
// @Param body body object true "Device Serial"
// @Success 200 {string} string "OK"
// @Router /users/me/device [put]
func (h *MobileHandler) LinkDevice(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req struct {
		DeviceSerial string `json:"device_serial"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Update("device_serial", req.DeviceSerial).Error; err != nil {
		http.Error(w, "Failed to link device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateSchedule creates a new medication schedule
// @Summary Create schedule
// @Tags Mobile
// @Security MobileAuth
// @Accept json
// @Param body body models.Schedule true "Schedule details"
// @Success 201 {object} models.Schedule
// @Router /schedules [post]
func (h *MobileHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var schedule models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	schedule.UserID = userID

	if err := h.DB.Create(&schedule).Error; err != nil {
		http.Error(w, "Failed to create schedule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// ListSchedules lists all schedules for the user
// @Summary List schedules
// @Tags Mobile
// @Security MobileAuth
// @Produce json
// @Success 200 {array} models.Schedule
// @Router /schedules [get]
func (h *MobileHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var schedules []models.Schedule
	if err := h.DB.Where("user_id = ?", userID).Find(&schedules).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// DeleteSchedule deletes a schedule
// @Summary Delete schedule
// @Tags Mobile
// @Security MobileAuth
// @Param id path int true "Schedule ID"
// @Success 204 {string} string "No Content"
// @Router /schedules/{id} [delete]
func (h *MobileHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	id := chi.URLParam(r, "id")

	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Schedule{}).Error; err != nil {
		http.Error(w, "Failed to delete schedule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddCaregiver adds a caregiver
// @Summary Add caregiver
// @Tags Mobile
// @Security MobileAuth
// @Accept json
// @Param body body models.Caregiver true "Caregiver details"
// @Success 201 {object} models.Caregiver
// @Router /caregivers [post]
func (h *MobileHandler) AddCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var caregiver models.Caregiver
	if err := json.NewDecoder(r.Body).Decode(&caregiver); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	caregiver.PatientID = userID

	if err := h.DB.Create(&caregiver).Error; err != nil {
		http.Error(w, "Failed to add caregiver", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(caregiver)
}

// ListCaregivers lists all caregivers
// @Summary List caregivers
// @Tags Mobile
// @Security MobileAuth
// @Produce json
// @Success 200 {array} models.Caregiver
// @Router /caregivers [get]
func (h *MobileHandler) ListCaregivers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var caregivers []models.Caregiver
	if err := h.DB.Where("patient_id = ?", userID).Find(&caregivers).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(caregivers)
}

// DeleteCaregiver deletes a caregiver
// @Summary Delete caregiver
// @Tags Mobile
// @Security MobileAuth
// @Param id path int true "Caregiver ID"
// @Success 204 {string} string "No Content"
// @Router /caregivers/{id} [delete]
func (h *MobileHandler) DeleteCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	id := chi.URLParam(r, "id")

	if err := h.DB.Where("id = ? AND patient_id = ?", id, userID).Delete(&models.Caregiver{}).Error; err != nil {
		http.Error(w, "Failed to delete caregiver", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RegisterPushToken registers a mobile device push token
// @Summary Register push token
// @Tags Mobile
// @Security MobileAuth
// @Accept json
// @Param body body models.PushToken true "Push token details"
// @Success 201 {string} string "Created"
// @Router /push-tokens [post]
func (h *MobileHandler) RegisterPushToken(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var token models.PushToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	token.UserID = userID

	// Upsert to handle unique tokens being reused/re-registered
	if err := h.DB.Save(&token).Error; err != nil {
		http.Error(w, "Failed to register push token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ListIntakeLogs retrieves intake history
// @Summary List intake logs
// @Tags Mobile
// @Security MobileAuth
// @Produce json
// @Success 200 {array} models.IntakeLog
// @Router /intake-logs [get]
func (h *MobileHandler) ListIntakeLogs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var logs []models.IntakeLog
	
	// Optional filtering by date could be added here parsing query params
	if err := h.DB.Where("user_id = ?", userID).Find(&logs).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
