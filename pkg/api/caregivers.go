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

// CaregiverRequest is the payload for setting caregivers
type CaregiverRequest struct {
	Caregivers []CaregiverInput `json:"caregivers"`
}

// VerifyPhoneRequest is the payload for verifying a caregiver phone OTP
type VerifyPhoneRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

// SetCaregivers replaces all caregivers for the user
// @Summary Set caregivers (full replace)
// @Tags Caregivers
// @Security MobileAuth
// @Accept json
// @Produce json
// @Param body body CaregiverRequest true "List of caregivers"
// @Success 200 {array} CaregiverInput
// @Router /api/v1/caregivers [put]
func (h *MobileHandler) SetCaregivers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req CaregiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Delete all existing caregivers for this user
	if err := h.DB.Where("patient_id = ?", userID).Delete(&models.Caregiver{}).Error; err != nil {
		http.Error(w, "Failed to clear caregivers", http.StatusInternalServerError)
		return
	}

	var result []CaregiverInput
	for _, input := range req.Caregivers {
		if input.Email == "" && input.Phone == "" {
			continue
		}

		cg := models.Caregiver{PatientID: userID, Email: input.Email, Phone: input.Phone}
		if err := h.DB.Create(&cg).Error; err != nil {
			http.Error(w, "Failed to add caregiver", http.StatusInternalServerError)
			return
		}
		result = append(result, CaregiverInput{Email: cg.Email, Phone: cg.Phone})

		if input.Email != "" {
			email.VerifyEmail(input.Email)
		}
		if input.Phone != "" {
			sms.SendSandboxOTP(input.Phone)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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
// @Success 200 {array} CaregiverInput
// @Router /api/v1/caregivers [get]
func (h *MobileHandler) ListCaregivers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var caregivers []models.Caregiver
	if err := h.DB.Where("patient_id = ?", userID).Find(&caregivers).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	result := make([]CaregiverInput, len(caregivers))
	for i, cg := range caregivers {
		result[i] = CaregiverInput{Email: cg.Email, Phone: cg.Phone}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
