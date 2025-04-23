package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/gohost/internal/auth"
)

type loginRequest struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type loginResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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

	// Login successful
	response := loginResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
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
