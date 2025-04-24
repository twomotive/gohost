package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/twomotive/gohost/internal/auth" // Import auth package
)

type stripWebhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handleStripWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- API Key Verification Start ---
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("Error getting API key from webhook request: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if apiKey != cfg.stripKey {
		log.Printf("Invalid API key received for webhook: %s", apiKey)
		http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
		return
	}
	// --- API Key Verification End ---

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req stripWebhookRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON webhook decode error: %v", err)
		http.Error(w, "Invalid request body: expected JSON format", http.StatusBadRequest)
		return
	}

	// Only process 'user.upgraded' events
	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Validate and parse the user ID
	userID, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		log.Printf("Invalid user ID format in webhook: %v", err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Update the user's membership status in the database
	_, err = cfg.db.UpdateUserMembership(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found for webhook upgrade: %s", userID)
			http.Error(w, "User not found", http.StatusNotFound) // 404
		} else {
			log.Printf("Database error updating user membership via webhook for user %s: %v", userID, err)
			// Return 500 for internal errors, strip should retry
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("User %s successfully upgraded to Gohost Red via webhook.", userID)
	w.WriteHeader(http.StatusNoContent) // 204 - Success
}
