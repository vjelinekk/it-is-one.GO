package api

import (
	"encoding/json"
	"net/http"

	"github.com/vjelinekk/it-is-one.GO/pkg/models"
)

// DeletedIDResponse contains the ID of a deleted resource
type DeletedIDResponse struct {
	ID uint `json:"id"`
}

// SetScheduleRequest is the payload for setting schedules
type SetScheduleRequest struct {
	Time1 string `json:"time1"` // e.g. '08:00:00'
	Time2 string `json:"time2"` // e.g. '20:00:00'
}

// SetSchedules replaces all schedules for the user with time1 and time2
// @Summary Set schedules (full replace)
// @Tags Schedules
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body SetScheduleRequest true "Schedule times"
// @Success 200 {object} SetScheduleRequest
// @Router /api/v1/schedules [put]
func (h *MobileHandler) SetSchedules(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req SetScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if req.Time1 == "" || req.Time2 == "" {
		http.Error(w, "Both time1 and time2 are required", http.StatusBadRequest)
		return
	}

	if req.Time1 == req.Time2 {
		http.Error(w, "time1 and time2 must be different", http.StatusBadRequest)
		return
	}

	// Delete all existing schedules for this user
	if err := h.DB.Where("user_id = ?", userID).Delete(&models.Schedule{}).Error; err != nil {
		http.Error(w, "Failed to clear schedules", http.StatusInternalServerError)
		return
	}

	for _, t := range []string{req.Time1, req.Time2} {
		s := models.Schedule{UserID: userID, ScheduledTime: t}
		if err := h.DB.Create(&s).Error; err != nil {
			http.Error(w, "Failed to create schedule", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// ListSchedules lists all schedules for the user
// @Summary List schedules
// @Tags Schedules
// @Security MobileAuth
// @Produce json
// @Success 200 {object} SetScheduleRequest
// @Router /api/v1/schedules [get]
func (h *MobileHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var schedules []models.Schedule
	if err := h.DB.Where("user_id = ?", userID).Order("scheduled_time ASC").Find(&schedules).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	result := SetScheduleRequest{}
	if len(schedules) > 0 {
		result.Time1 = schedules[0].ScheduledTime
	}
	if len(schedules) > 1 {
		result.Time2 = schedules[1].ScheduledTime
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
