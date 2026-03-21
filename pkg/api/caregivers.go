package api

import (
	"encoding/json"
	"net/http"

	"github.com/vjelinekk/it-is-one.GO/pkg/email"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
)

// CaregiverRequest is the payload for adding caregivers
type CaregiverRequest struct {
	Emails []string `json:"emails"`
}

// DeletedIDResponse contains the ID of a deleted resource
type DeletedIDResponse struct {
	ID uint `json:"id"`
}

// AddCaregivers adds one or more caregivers
// @Summary Add caregivers
// @Tags Caregivers
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body CaregiverRequest true "List of caregiver emails"
// @Success 201 {array} CaregiverRequest
// @Router /api/v1/caregivers [post]
func (h *MobileHandler) AddCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if len(req.Emails) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	caregivers := make([]models.Caregiver, len(req.Emails))
	for i, email := range req.Emails {
		caregivers[i] = models.Caregiver{PatientID: userID, Email: email}
	}
	if err := h.DB.Create(&caregivers).Error; err != nil {
		http.Error(w, "Failed to add caregivers", http.StatusInternalServerError)
		return
	}

	result := make([]string, len(caregivers))
	for i, c := range caregivers {
		result[i] = c.Email
		email.VerifyEmail(c.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CaregiverRequest{Emails: result})
}

// ListCaregivers lists all caregivers
// @Summary List caregivers
// @Tags Caregivers
// @Security MobileAuth
// @Produce json
// @Success 200 {array} CaregiverRequest
// @Router /api/v1/caregivers [get]
func (h *MobileHandler) ListCaregivers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var caregivers []models.Caregiver
	if err := h.DB.Where("patient_id = ?", userID).Find(&caregivers).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	result := make([]string, len(caregivers))
	for i, c := range caregivers {
		result[i] = c.Email
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DeleteCaregiver deletes a caregiver
// @Summary Delete caregivers
// @Tags Caregivers
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body CaregiverRequest true "List of caregiver emails"
// @Success 200 {object} CaregiverRequest
// @Router /api/v1/caregivers [delete]
func (h *MobileHandler) DeleteCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if len(req.Emails) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CaregiverRequest{Emails: []string{}})
		return
	}

	if err := h.DB.Where("patient_id = ? AND email IN ?", userID, req.Emails).Delete(&models.Caregiver{}).Error; err != nil {
		http.Error(w, "Failed to delete caregiver", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CaregiverRequest{Emails: req.Emails})
}
