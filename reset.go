package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileServerHits.Store(0)
	fmt.Fprintf(w, "Hits has been reset to %d", cfg.fileServerHits.Load())

}
