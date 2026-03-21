package api

import (
	"encoding/json"
	"net/http"

	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"gorm.io/gorm"
)

// UserHandler handles user-related requests
type UserHandler struct {
	DB *gorm.DB
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// CreateUserRequest is the payload for creating a user
type CreateUserRequest struct {
	Email string `json:"email"`
}

// CreateUserResponse is the response for creating a user
type CreateUserResponse struct {
	ID uint `json:"id"`
}

// Create handles creating a new user
// @Summary Create user
// @Tags Users
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "User email"
// @Success 201 {object} CreateUserResponse
// @Router /api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user := models.User{Email: req.Email}
	if err := h.DB.Create(&user).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateUserResponse{ID: user.ID})
}

// PatchUserRequest is the payload for patching user notification settings
type PatchUserRequest struct {
	NotifyAfterMinutes           *int `json:"notify_after_minutes"`
	NotifyCaregiversAfterRetries *int `json:"notify_caregivers_after_retries"`
}

// Patch updates the authenticated user's notification settings
// @Summary Patch user notification settings
// @Tags Users
// @Security MobileAuth
// @Accept json
// @Param body body PatchUserRequest true "Notification settings"
// @Success 204
// @Router /api/v1/users [patch]
func (h *UserHandler) Patch(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(uint)
	var req PatchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})
	if req.NotifyAfterMinutes != nil {
		updates["notify_after_minutes"] = *req.NotifyAfterMinutes
	}
	if req.NotifyCaregiversAfterRetries != nil {
		updates["notify_caregivers_after_retries"] = *req.NotifyCaregiversAfterRetries
	}

	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
