package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/twomotive/gohost/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	jwtSecret      string
	stripKey       string // Add stripKey field
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set!!")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set!!")
	}

	stripKey := os.Getenv("STRIP_KEY") // Load STRIP_KEY
	if stripKey == "" {
		log.Fatal("STRIP_KEY must be set!!")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}

	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		jwtSecret:      jwtSecret,
		stripKey:       stripKey, // Store stripKey in config
	}

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

	// Add reset endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	// Add validate endpoint
	mux.HandleFunc("POST /api/validate", handleValidate)

	// Add users api endpoint to create users
	mux.HandleFunc("POST /api/users", apiCfg.createUsers)

	mux.HandleFunc("PUT /api/users", apiCfg.updateUser)

	mux.HandleFunc("POST /api/login", apiCfg.userLogin)

	// Add refresh token endpoint
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)

	// Add revoke token endpoint
	mux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)

	mux.HandleFunc("POST /api/gobits", apiCfg.createGoBits)

	mux.HandleFunc("GET /api/gobits", apiCfg.getAllGoBits)

	mux.HandleFunc("GET /api/gobits/{gobitID}", apiCfg.getGoBitByID)

	mux.HandleFunc("DELETE /api/gobits/{gobitID}", apiCfg.deleteGobit)

	mux.HandleFunc("POST /api/strip/webhooks", apiCfg.handleStripWebhook)

	fmt.Println("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\\n", err)
	}
}
