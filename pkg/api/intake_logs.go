package api

import (
	"net/http"
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

// LogIntake records that the user took their medication
// @Summary Log medication intake
// @Tags IntakeLogs
// @Security MobileAuth
// @Success 204
// @Router /api/v1/intake-logs [post]
func (h *IntakeLogHandler) LogIntake(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)

	log := models.IntakeLog{
		UserID:  userID,
		TakenAt: time.Now(),
	}
	if err := h.DB.Create(&log).Error; err != nil {
		http.Error(w, "Failed to log intake", http.StatusInternalServerError)
		return
	}

	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Update("current_missed_doses", 0).Error; err != nil {
		http.Error(w, "Failed to reset missed doses", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
