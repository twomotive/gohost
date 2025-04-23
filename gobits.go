package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/twomotive/gohost/internal/database"
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

func (cfg *apiConfig) handleGoBits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req gobitRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("JSON gobit decode error: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Body == "" {
		http.Error(w, "Body cannot be empty", http.StatusBadRequest)
		return
	}

	if req.UserID == uuid.Nil {
		http.Error(w, "Invalid UserID", http.StatusBadRequest)
		return
	}
	params := database.CreateGobitParams{
		Body:   req.Body,
		UserID: req.UserID,
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
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 for gobit creation
	w.Write(data)
}
