package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sort" // Import the sort package
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/gohost/internal/auth"
	"github.com/twomotive/gohost/internal/database" // Ensure database import is present
)

type gobitRequest struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type createdGobit struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createGoBits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- Authentication Start ---
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error validating JWT: %v", err)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}
	// --- Authentication End ---

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req gobitRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON gobit decode error: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Body == "" {
		http.Error(w, "Body cannot be empty", http.StatusBadRequest)
		return
	}

	params := database.CreateGobitParams{
		Body:   req.Body,
		UserID: userID,
	}

	gobit, err := cfg.db.CreateGobit(r.Context(), params)
	if err != nil {
		log.Printf("cannot create gobit !!: %v", err)
		http.Error(w, "Failed to create gobit", http.StatusInternalServerError)
		return
	}

	responseGobit := createdGobit{
		ID:        gobit.ID,
		CreatedAt: gobit.CreatedAt,
		UpdatedAt: gobit.UpdatedAt,
		Body:      gobit.Body,
		UserID:    gobit.UserID,
	}

	data, err := json.Marshal(responseGobit)
	if err != nil {
		log.Printf("Error marshalling gobit response: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 for gobit creation
	w.Write(data)
}

func (cfg *apiConfig) getAllGoBits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authorIDStr := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort") // Get the sort parameter
	var gobits []database.Gobit
	var err error

	if authorIDStr != "" {
		// If author_id is provided, parse it and get gobits by author
		authorID, parseErr := uuid.Parse(authorIDStr)
		if parseErr != nil {
			log.Printf("Invalid author_id format: %v", parseErr)
			http.Error(w, "Invalid author_id format", http.StatusBadRequest)
			return
		}
		gobits, err = cfg.db.GetGobitsByAuthor(r.Context(), authorID)
	} else {
		// If author_id is not provided, get all gobits
		gobits, err = cfg.db.GetAllGobits(r.Context())
	}

	if err != nil {
		log.Printf("cannot get gobits: %v", err)
		http.Error(w, "Failed to get gobits", http.StatusInternalServerError)
		return
	}

	// Sort the gobits based on the sort parameter if it's "desc"
	if sortParam == "desc" {
		sort.Slice(gobits, func(i, j int) bool {
			// For descending order, check if i's time is after j's time
			return gobits[i].CreatedAt.After(gobits[j].CreatedAt)
		})
	}
	// Default is ascending, which is already handled by the SQL query.

	responseGobits := make([]createdGobit, len(gobits))
	for i, dbGobit := range gobits {
		responseGobits[i] = createdGobit{
			ID:        dbGobit.ID,
			CreatedAt: dbGobit.CreatedAt,
			UpdatedAt: dbGobit.UpdatedAt,
			Body:      dbGobit.Body,
			UserID:    dbGobit.UserID,
		}
	}

	data, err := json.Marshal(responseGobits)
	if err != nil {
		log.Printf("Error marshalling gobits response: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK for successful retrieval
	w.Write(data)

}

func (cfg *apiConfig) getGoBitByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract gobitID from the URL path parameter
	gobitIDStr := r.PathValue("gobitID")
	if gobitIDStr == "" {
		http.Error(w, "gobit ID is required", http.StatusBadRequest)
		return
	}

	gobitID, err := uuid.Parse(gobitIDStr)
	if err != nil {
		http.Error(w, "Invalid gobit ID format", http.StatusBadRequest)
		return
	}

	dbGobit, err := cfg.db.GetGobit(r.Context(), gobitID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "gobit not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting gobit by ID %s: %v", gobitID, err)
			http.Error(w, "Failed to get gobit", http.StatusInternalServerError)
		}
		return
	}

	responseGobit := createdGobit{
		ID:        dbGobit.ID,
		CreatedAt: dbGobit.CreatedAt,
		UpdatedAt: dbGobit.UpdatedAt,
		Body:      dbGobit.Body,
		UserID:    dbGobit.UserID,
	}

	data, err := json.Marshal(responseGobit)
	if err != nil {
		log.Printf("Error marshalling single gobit response: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (cfg *apiConfig) deleteGobit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- Authentication Start ---
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token for delete: %v", err)
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		log.Printf("Error validating JWT for delete: %v", err)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}
	// --- Authentication End ---

	// Extract gobitID from the URL path parameter
	gobitIDStr := r.PathValue("gobitID")
	if gobitIDStr == "" {
		http.Error(w, "gobit ID is required", http.StatusBadRequest)
		return
	}

	gobitID, err := uuid.Parse(gobitIDStr)
	if err != nil {
		http.Error(w, "Invalid gobit ID format", http.StatusBadRequest)
		return
	}

	// First, get the gobit to check ownership
	dbGobit, err := cfg.db.GetGobit(r.Context(), gobitID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "gobit not found", http.StatusNotFound)
		} else {
			log.Printf("Error getting gobit %s for deletion check: %v", gobitID, err)
			http.Error(w, "Failed to get gobit", http.StatusInternalServerError)
		}
		return
	}

	// Check if the authenticated user is the author
	if dbGobit.UserID != userID {
		log.Printf("User %s attempted to delete gobit %s owned by user %s", userID, gobitID, dbGobit.UserID)
		http.Error(w, "Forbidden: You do not own this gobit", http.StatusForbidden)
		return
	}

	// User is the author, proceed with deletion
	err = cfg.db.DeleteGobit(r.Context(), database.DeleteGobitParams{
		ID:     gobitID,
		UserID: userID,
	})
	if err != nil {

		log.Printf("Error deleting gobit %s by user %s: %v", gobitID, userID, err)
		http.Error(w, "Failed to delete gobit", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
}
