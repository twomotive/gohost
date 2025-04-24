package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/gohost/internal/auth"
	"github.com/twomotive/gohost/internal/database"
)

type loginRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type loginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON login request decode error: %v", err)
		http.Error(w, "Invalid request body: expected JSON format", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found for email: %s", req.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			log.Printf("Database error getting user by email: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Check password hash
	err = auth.CheckPasswordHash(user.HashedPassword, req.Password)
	if err != nil {
		log.Printf("Password mismatch for user: %s", req.Email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	jwtExpiresIn := time.Hour

	// Generate JWT
	tokenString, err := auth.MakeJWT(user.ID, cfg.jwtSecret, jwtExpiresIn)
	if err != nil {
		log.Printf("Error generating JWT for user %s: %v", user.Email, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate Refresh Token
	refreshTokenString, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error generating refresh token for user %s: %v", user.Email, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	refreshTokenExpiresAt := time.Now().Add(60 * 24 * time.Hour) // 60 days expiration
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshTokenString,
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpiresAt,
	})
	if err != nil {
		log.Printf("Error storing refresh token for user %s: %v", user.Email, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Login successful
	response := loginResponse{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        tokenString,
		RefreshToken: refreshTokenString,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
