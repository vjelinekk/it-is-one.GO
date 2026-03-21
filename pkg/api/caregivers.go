package api

import (
	"encoding/json"
	"net/http"

	"github.com/vjelinekk/it-is-one.GO/pkg/email"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"github.com/vjelinekk/it-is-one.GO/pkg/sms"
)

// CaregiverInput is one caregiver entry with optional email and/or phone
type CaregiverInput struct {
	Email string `json:"email"`
	Phone string `json:"phone"` // E.164 format e.g. +420123456789
}

// CaregiverRequest is the payload for adding caregivers
type CaregiverRequest struct {
	Caregivers []CaregiverInput `json:"caregivers"`
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
// @Param body body CaregiverRequest true "List of caregivers"
// @Success 201 {array} models.Caregiver
// @Router /api/v1/caregivers [post]
func (h *MobileHandler) AddCaregiver(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	if len(req.Caregivers) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	var created []models.Caregiver
	for _, input := range req.Caregivers {
		if input.Email == "" && input.Phone == "" {
			continue
		}

		// Skip duplicates per patient
		var count int64
		query := h.DB.Model(&models.Caregiver{}).Where("patient_id = ?", userID)
		if input.Email != "" && input.Phone != "" {
			query = query.Where("email = ? OR phone = ?", input.Email, input.Phone)
		} else if input.Email != "" {
			query = query.Where("email = ?", input.Email)
		} else {
			query = query.Where("phone = ?", input.Phone)
		}
		query.Count(&count)
		if count > 0 {
			continue
		}

		cg := models.Caregiver{PatientID: userID, Email: input.Email, Phone: input.Phone}
		if err := h.DB.Create(&cg).Error; err != nil {
			http.Error(w, "Failed to add caregiver", http.StatusInternalServerError)
			return
		}
		created = append(created, cg)

		if input.Email != "" {
			email.VerifyEmail(input.Email)
		}
		if cg.Phone != "" {
			sms.SendSandboxOTP(cg.Phone)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// VerifyPhoneRequest is the payload for verifying a caregiver phone OTP
type VerifyPhoneRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

// VerifyPhone verifies a caregiver phone number OTP in SNS sandbox
// @Summary Verify caregiver phone OTP
// @Tags Caregivers
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body VerifyPhoneRequest true "Phone and OTP"
// @Success 200
// @Router /api/v1/caregivers/verify-phone [post]
func (h *MobileHandler) VerifyPhone(w http.ResponseWriter, r *http.Request) {
	var req VerifyPhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Phone == "" || req.OTP == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := sms.VerifySandboxOTP(req.Phone, req.OTP); err != nil {
		http.Error(w, "OTP verification failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ListCaregivers lists all caregivers
// @Summary List caregivers
// @Tags Caregivers
// @Security MobileAuth
// @Produce json
// @Success 200 {array} models.Caregiver
// @Router /api/v1/caregivers [get]
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
