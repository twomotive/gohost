package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	apiCfg := &apiConfig{}
	mux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Update fileserver paths with metrics middleware
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fileServer)))
	mux.Handle("/assets/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("assets"))))

	// Add health check endpoint
	mux.HandleFunc("GET /api/healthz", HandleReadiness)

	// Add metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)

	// // Add reset endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	fmt.Println("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
