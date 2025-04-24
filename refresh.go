package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/twomotive/gohost/internal/auth"
)

type refreshResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token for refresh: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Look up the refresh token in the database
	refreshTokenData, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		// Consider sql.ErrNoRows as unauthorized
		log.Printf("Error retrieving refresh token data: %v", err)
		http.Error(w, "Unauthorized: Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Check if the token has been revoked
	if refreshTokenData.RevokedAt.Valid {
		log.Printf("Attempt to use revoked refresh token for user %s", refreshTokenData.UserID)
		http.Error(w, "Unauthorized: Refresh token revoked", http.StatusUnauthorized)
		return
	}

	// Check if the token has expired
	if time.Now().After(refreshTokenData.ExpiresAt) {
		log.Printf("Attempt to use expired refresh token for user %s", refreshTokenData.UserID)
		http.Error(w, "Unauthorized: Refresh token expired", http.StatusUnauthorized)
		return
	}

	// Token is valid, issue a new access token
	newJwtExpiresIn := time.Hour
	newAccessTokenString, err := auth.MakeJWT(refreshTokenData.UserID, cfg.jwtSecret, newJwtExpiresIn)
	if err != nil {
		log.Printf("Error generating new JWT during refresh for user %s: %v", refreshTokenData.UserID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := refreshResponse{
		Token: newAccessTokenString,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling refresh response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
