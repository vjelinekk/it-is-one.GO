package api

import (
	"net/http"
	"sort"
	"time"

	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"gorm.io/gorm"
)

type IntakeLogHandler struct {
	DB *gorm.DB
}

func NewIntakeLogHandler(db *gorm.DB) *IntakeLogHandler {
	return &IntakeLogHandler{DB: db}
}

// LogIntake records that the user took their medication.
// It logs the earliest scheduled dose for today that has not been taken yet.
// @Summary Log medication intake
// @Tags IntakeLogs
// @Security MobileAuth
// @Success 204
// @Router /api/v1/intake-logs [post]
func (h *IntakeLogHandler) LogIntake(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)

	var user models.User
	if err := h.DB.Preload("Schedules").First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil || user.Timezone == "" {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	today := now.Format("2006-01-02")

	// Sort schedules by scheduled time so dose 1 < dose 2
	sort.Slice(user.Schedules, func(i, j int) bool {
		return user.Schedules[i].ScheduledTime < user.Schedules[j].ScheduledTime
	})

	// Find the first dose slot not yet taken today
	for i := range user.Schedules {
		doseSlot := i + 1
		var count int64
		h.DB.Model(&models.IntakeLog{}).
			Where("user_id = ? AND dose_slot = ? AND date = ?", userID, doseSlot, today).
			Count(&count)
		if count > 0 {
			continue
		}

		entry := models.IntakeLog{
			UserID:   userID,
			DoseSlot: doseSlot,
			Date:     today,
			TakenAt:  time.Now(),
		}
		if err := h.DB.Create(&entry).Error; err != nil {
			http.Error(w, "Failed to log intake", http.StatusInternalServerError)
			return
		}
		h.DB.Model(&models.User{}).Where("id = ?", userID).Update("current_missed_doses", 0)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// All doses already taken today
	w.WriteHeader(http.StatusNoContent)
}
