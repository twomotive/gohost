package main

import (
	"net/http"
	"os"
)

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileServerHits.Store(0)

	if os.Getenv("PLATFORM") != "dev" {
		http.Error(w, "not in development mode!!", http.StatusForbidden)
		return
	}

	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}
