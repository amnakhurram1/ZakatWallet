package main

// main.go boots the REST API server. It initializes a new
// blockchain with a genesis block paying to a hard-coded address,
// constructs the API server and listens on port 8080. All routes are
// versioned under /api/v1.

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"wallet_backend_go/internal/api"
	"wallet_backend_go/internal/blockchain"
)

// withCORS wraps the given handler and adds CORS headers so that
// the React frontend (running on http://localhost:3000) can call
// the Go API on http://localhost:8080 without being blocked.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow your frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		// If you want to be looser during dev, you *could* use "*"
		// w.Header().Set("Access-Control-Allow-Origin", "*")

		// Let proxies / caches know this varies by Origin
		w.Header().Set("Vary", "Origin")

		// Allowed methods and headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Normal request: pass to the router
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables from .env (if present)
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	// Create a new blockchain with a dummy genesis recipient. In a
	// real deployment you might take this from config or an env var.
	bc := blockchain.NewBlockchain("b2185e5380ecc4f928877552981268dbc04836b6d44942cca8a3e60a29af2211")
	srv := api.NewServer(bc)

	// Wrap the router with CORS middleware
	handler := withCORS(srv.Router())

	log.Println("Starting blockchain wallet backend on port 8080…")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
