package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"gorm.io/gorm"
)

var errScheduleTimeAlreadyExists = errors.New("schedule time already exists")

// DeletedIDResponse contains the ID of a deleted resource
type DeletedIDResponse struct {
	ID uint `json:"id"`
}

// CreateScheduleRequest is the payload for creating schedules
type CreateScheduleRequest struct {
	Time1 string `json:"time1"` // e.g. '08:00:00'
	Time2 string `json:"time2"` // e.g. '20:00:00'
}

// PatchScheduleRequest is the payload for patching schedule times
type PatchScheduleRequest struct {
	Time1 *string `json:"time1"` // lower schedule ID
	Time2 *string `json:"time2"` // higher schedule ID
}

// CreateSchedule creates schedule rows for time1 and time2, skipping duplicates
// @Summary Create schedules
// @Tags Schedules
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body CreateScheduleRequest true "Schedule times"
// @Success 201 {array} models.Schedule
// @Router /api/v1/schedules [post]
func (h *MobileHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var created []models.Schedule
	for _, t := range []string{req.Time1, req.Time2} {
		if t == "" {
			continue
		}
		var count int64
		h.DB.Model(&models.Schedule{}).
			Where("user_id = ? AND scheduled_time = ?", userID, t).
			Count(&count)
		if count > 0 {
			continue
		}
		s := models.Schedule{UserID: userID, ScheduledTime: t}
		if err := h.DB.Create(&s).Error; err != nil {
			http.Error(w, "Failed to create schedule", http.StatusInternalServerError)
			return
		}
		created = append(created, s)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// ListSchedules lists all schedules for the user
// @Summary List schedules
// @Tags Schedules
// @Security MobileAuth
// @Produce json
// @Success 200 {array} models.Schedule
// @Router /api/v1/schedules [get]
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

// PatchSchedule updates schedule rows for time1/time2 based on schedule ID ordering
// @Summary Patch schedules
// @Tags Schedules
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body PatchScheduleRequest true "Schedule times"
// @Success 200 {array} models.Schedule
// @Router /api/v1/schedules [patch]
func (h *MobileHandler) PatchSchedule(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req PatchScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	updateTime1 := req.Time1 != nil && *req.Time1 != ""
	updateTime2 := req.Time2 != nil && *req.Time2 != ""

	if !updateTime1 && !updateTime2 {
		http.Error(w, "At least one of time1 or time2 is required", http.StatusBadRequest)
		return
	}

	if updateTime1 && updateTime2 && *req.Time1 == *req.Time2 {
		http.Error(w, "time1 and time2 must be different", http.StatusBadRequest)
		return
	}

	var schedules []models.Schedule
	if err := h.DB.Where("user_id = ?", userID).Order("id ASC").Find(&schedules).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(schedules) == 0 {
		http.Error(w, "No schedules found", http.StatusNotFound)
		return
	}

	if updateTime2 && len(schedules) < 2 {
		http.Error(w, "No schedule found for time2", http.StatusNotFound)
		return
	}

	updated := make([]models.Schedule, 0, 2)
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if updateTime1 {
			target := schedules[0]

			var count int64
			if err := tx.Model(&models.Schedule{}).
				Where("user_id = ? AND scheduled_time = ? AND id <> ?", userID, *req.Time1, target.ID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errScheduleTimeAlreadyExists
			}

			if err := tx.Model(&models.Schedule{}).
				Where("id = ? AND user_id = ?", target.ID, userID).
				Update("scheduled_time", *req.Time1).Error; err != nil {
				return err
			}

			target.ScheduledTime = *req.Time1
			updated = append(updated, target)
		}

		if updateTime2 {
			target := schedules[1]

			var count int64
			if err := tx.Model(&models.Schedule{}).
				Where("user_id = ? AND scheduled_time = ? AND id <> ?", userID, *req.Time2, target.ID).
				Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errScheduleTimeAlreadyExists
			}

			if err := tx.Model(&models.Schedule{}).
				Where("id = ? AND user_id = ?", target.ID, userID).
				Update("scheduled_time", *req.Time2).Error; err != nil {
				return err
			}

			target.ScheduledTime = *req.Time2
			updated = append(updated, target)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, errScheduleTimeAlreadyExists) {
			http.Error(w, "Schedule time already exists", http.StatusBadRequest)
			return
		}
		http.Error(w, "Failed to update schedules", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// DeleteSchedule deletes a schedule
// @Summary Delete schedule
// @Tags Schedules
// @Security MobileAuth
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {object} DeletedIDResponse
// @Router /api/v1/schedules/{id} [delete]
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
