package main

import (
	"log"
	"net/http"
	"os"
	"pos-backend/internal/api"
	"pos-backend/internal/engine"
	"pos-backend/internal/repository"
)

func main() {
	log.Println("Initializing POS Backend Engine...")

	// 1. Initialize Database Connection
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/pos?sslmode=disable"
		log.Println("DATABASE_URL not set, using default local string")
	}

	repo, err := repository.NewRepository(connStr)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v. Persistence will be disabled.", err)
	}

	// 2. Initialize the Engine
	e := engine.NewEngine(repo)

	// 3. Boot Sequence: Rebuild state from logs
	if repo != nil {
		if err := e.RebuildState(); err != nil {
			log.Printf("Warning: Failed to rebuild state: %v", err)
		}
	}

	// 4. Start the Background Worker
	e.StartWorker()

	// 5. Initialize the API Server
	server := api.NewServer(e)

	// 4. Setup Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	mux.HandleFunc("/api/ingest", server.IngestHandler)
	mux.HandleFunc("/api/search", server.SearchHandler)
	mux.HandleFunc("/api/undo", server.UndoHandler)
	mux.HandleFunc("/api/sort", server.SortHandler)
	mux.HandleFunc("/api/state", server.StateHandler)

	// Add CORS middleware
	handler := corsMiddleware(mux)

	// 5. Start the HTTP Server
	port := ":8080"
	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
