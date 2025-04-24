package main

import (
	"log"
	"net/http"

	"github.com/twomotive/gohost/internal/auth"
)

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error getting bearer token for revoke: %v", err)
		// Respond with 204 even if the token is invalid or missing, as per common practice for revocation
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Attempt to revoke the token in the database
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		// Log the error but still return 204. The client doesn't need to know if the token existed or if there was a DB error.
		log.Printf("Error revoking refresh token: %v", err)
	}

	// Respond with 204 No Content regardless of whether the token was found and revoked
	w.WriteHeader(http.StatusNoContent)
}
