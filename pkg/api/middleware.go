package api

import (
	"context"
	"net/http"
	"strconv"
)

type contextKey string

const (
	UserIDKey       contextKey = "user_id"
	DeviceSerialKey contextKey = "device_serial"
)

// MobileAuthMiddleware extracts the X-User-ID header and adds it to the context
func MobileAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Error(w, "Missing X-User-ID header", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid X-User-ID format", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, uint(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HardwareAuthMiddleware extracts the X-Device-Serial header and adds it to the context
func HardwareAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serial := r.Header.Get("X-Device-Serial")
		if serial == "" {
			http.Error(w, "Missing X-Device-Serial header", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), DeviceSerialKey, serial)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HeartbeatAuthMiddleware accepts either X-Device-Serial or X-User-ID
func HeartbeatAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if serial := r.Header.Get("X-Device-Serial"); serial != "" {
			ctx := context.WithValue(r.Context(), DeviceSerialKey, serial)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Error(w, "Missing X-Device-Serial or X-User-ID header", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid X-User-ID format", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, uint(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
