package main

import (
	"log"
	"net/http"
	"os"
	"pos-backend/internal/api"
	"pos-backend/internal/engine"
	"pos-backend/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Initializing POS Backend Engine...")

	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		cwd, _ := os.Getwd()
		log.Printf("Warning: godotenv failed to load .env: %v (Current working directory: %s)", err, cwd)
	}

	// Get Supabase credentials for API-based repository
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseAnonKey == "" {
		log.Fatal("CRITICAL: SUPABASE_URL or SUPABASE_ANON_KEY not set. Backend requires persistence to operate.")
	}

	repo, err := repository.NewRepository(supabaseURL, supabaseAnonKey)
	if err != nil {
		log.Fatalf("CRITICAL: Could not initialize repository: %v", err)
	}

	// Strict connection check
	if err := repo.Verify(); err != nil {
		log.Fatalf("CRITICAL: Database connection verification failed: %v. Please check your network, credentials, and Supabase RLS policies.", err)
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

	// Middleware for CORS
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			// Allow the specific Vercel production domain and localhost for development
			if origin == "https://vento-five.vercel.app" || origin == "http://localhost:5000" || origin == "http://localhost:8080" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Fallback to production domain for browsers that don't send Origin in same-site
				w.Header().Set("Access-Control-Allow-Origin", "https://vento-five.vercel.app")
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Helper to chain middleware
	chain := func(h http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
		for _, m := range middlewares {
			h = m(h)
		}
		return h
	}

	// 4. Setup Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Protected API Routes
	http.HandleFunc("/api/ingest", chain(server.IngestHandler, api.AuthMiddleware, corsMiddleware))
	http.HandleFunc("/api/search", chain(server.SearchHandler, api.AuthMiddleware, corsMiddleware))
	http.HandleFunc("/api/undo", chain(server.UndoHandler, api.AuthMiddleware, corsMiddleware))
	http.HandleFunc("/api/sort", chain(server.SortHandler, api.AuthMiddleware, corsMiddleware))
	http.HandleFunc("/api/state", chain(server.StateHandler, api.AuthMiddleware, corsMiddleware))

	// 5. Start the HTTP Server
	port := ":8080"
	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
