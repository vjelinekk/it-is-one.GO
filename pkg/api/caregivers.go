package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
)

// CaregiverRequest is the payload for adding a caregiver
type CaregiverRequest struct {
	Email string `json:"email"`
}

// DeletedIDResponse contains the ID of a deleted resource
type DeletedIDResponse struct {
	ID uint `json:"id"`
}

// AddCaregiver adds a caregiver
// @Summary Add caregiver
// @Tags Caregivers
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body CaregiverRequest true "Caregiver details"
// @Success 201 {object} CaregiverRequest
// @Router /caregivers [post]
func (h *MobileHandler) AddCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	caregiver := models.Caregiver{
		PatientID: userID,
		Email:     req.Email,
	}
	if err := h.DB.Create(&caregiver).Error; err != nil {
		http.Error(w, "Failed to add caregiver", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CaregiverRequest{Email: caregiver.Email})
}

// ListCaregivers lists all caregivers
// @Summary List caregivers
// @Tags Caregivers
// @Security MobileAuth
// @Produce json
// @Success 200 {array} CaregiverRequest
// @Router /caregivers [get]
func (h *MobileHandler) ListCaregivers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var caregivers []models.Caregiver
	if err := h.DB.Where("patient_id = ?", userID).Find(&caregivers).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	result := make([]CaregiverRequest, len(caregivers))
	for i, c := range caregivers {
		result[i] = CaregiverRequest{Email: c.Email}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// DeleteCaregiver deletes a caregiver
// @Summary Delete caregiver
// @Tags Caregivers
// @Security MobileAuth
// @Produce json
// @Param id path int true "Caregiver ID"
// @Success 200 {object} DeletedIDResponse
// @Router /caregivers/{id} [delete]
func (h *MobileHandler) DeleteCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.Where("id = ? AND patient_id = ?", id, userID).Delete(&models.Caregiver{}).Error; err != nil {
		http.Error(w, "Failed to delete caregiver", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeletedIDResponse{ID: uint(id)})
}
