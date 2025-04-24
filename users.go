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
	Email    string `json:"email"`
	Password string `json:"password"`
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
	// Add check for empty password
	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("cannot hash password: %v", err)
		http.Error(w, "Internal server error processing request", http.StatusInternalServerError)
		return
	}

	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword, // Store the correctly hashed password
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

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- Authentication Start ---
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token for update: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error validating JWT for update: %v", err)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}
	// --- Authentication End ---

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req userRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON user update decode error: %v", err)
		http.Error(w, "Invalid request body: expected JSON format", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("cannot hash password during update: %v", err)
		http.Error(w, "Internal server error processing request", http.StatusInternalServerError)
		return
	}

	params := database.UpdateUserParams{
		ID:             userID,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.UpdateUser(r.Context(), params)
	if err != nil {

		log.Printf("cannot update user %s: %v", userID, err)
		http.Error(w, "Internal server error updating user", http.StatusInternalServerError)
		return
	}

	// Respond with the updated user data (excluding password)
	responseUser := createdUser{ // Re-use createdUser struct for response
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	data, err := json.Marshal(responseUser)
	if err != nil {
		log.Printf("Error marshalling update response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK for successful update
	w.Write(data)
}
