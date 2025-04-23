package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/gohost/internal/auth"
	"github.com/twomotive/gohost/internal/database"
)

type userRequest struct {
	Email          string `json:"email"`
	HashedPassword string `json:"hashed_password"`
}

type createdUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req userRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON user decode error: %v", err)
		http.Error(w, "Invalid request body: expected JSON format", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(req.HashedPassword)
	if err != nil {
		log.Printf("cannot hash password: %v", err)
		http.Error(w, "Internal server error processing request", http.StatusInternalServerError)
		return
	}

	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		log.Printf("cannot create user: %v", err)
		http.Error(w, "Internal server error creating user", http.StatusInternalServerError)
		return
	}

	responseUser := createdUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	data, err := json.Marshal(responseUser)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 for user creation
	w.Write(data)

}
