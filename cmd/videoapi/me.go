package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/warpcomdev/videoapi/internal/auth"
	"github.com/warpcomdev/videoapi/internal/models"
)

// response to /me enpoint
type meResponse struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Role    models.Role `json:"role"`
	Expires time.Time   `json:"expires"`
}

// handle calls to /me endpoint
func handleMe(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFrom(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := meResponse{
		ID:      claims.Subject,
		Name:    claims.Name,
		Role:    claims.Role,
		Expires: claims.ExpiresAt.Time,
	}
	json.NewEncoder(w).Encode(resp)
}
