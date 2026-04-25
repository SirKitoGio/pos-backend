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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/api/ingest", server.IngestHandler)
	http.HandleFunc("/api/search", server.SearchHandler)
	http.HandleFunc("/api/undo", server.UndoHandler)
	http.HandleFunc("/api/sort", server.SortHandler)
	http.HandleFunc("/api/state", server.StateHandler)

	// 5. Start the HTTP Server
	port := ":8080"
	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
