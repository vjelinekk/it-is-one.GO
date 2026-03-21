package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
)

// CreateScheduleRequest is the payload for creating a schedule
type CreateScheduleRequest struct {
	Time1 string `json:"time1"` // e.g. '08:00:00'
	Time2 string `json:"time2"` // e.g. '20:00:00'
}

// CreateSchedule creates a medication schedule with two times
// @Summary Create schedule
// @Tags Schedules
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body CreateScheduleRequest true "Schedule times"
// @Success 201 {object} models.Schedule
// @Router /schedules [post]
func (h *MobileHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	schedule := models.Schedule{
		UserID: userID,
		Time1:  req.Time1,
		Time2:  req.Time2,
	}
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
// @Tags Schedules
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
// @Tags Schedules
// @Security MobileAuth
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} DeletedIDResponse
// @Router /schedules/{id} [delete]
func (h *MobileHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Schedule{}).Error; err != nil {
		http.Error(w, "Failed to delete schedule", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeletedIDResponse{ID: uint(id)})
}
