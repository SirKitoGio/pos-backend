package api

import (
	"encoding/json"
	"net/http"
	"pos-backend/internal/engine"
	"pos-backend/internal/models"
)

type Server struct {
	Engine *engine.Engine
}

func NewServer(e *engine.Engine) *Server {
	return &Server{Engine: e}
}

func (s *Server) IngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tx models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	s.Engine.Queue <- tx
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
}

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	items := s.Engine.BST.SearchPrefix(query)
	if len(items) == 0 {
		http.Error(w, "No items found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (s *Server) UndoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	action, ok := s.Engine.Undo()
	if !ok {
		http.Error(w, "No actions to undo", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(action)
}

func (s *Server) StateHandler(w http.ResponseWriter, r *http.Request) {
	state := s.Engine.Matrix.GetState()
	queueSize := len(s.Engine.Queue)
	history := s.Engine.Stack.GetActions()

	response := map[string]interface{}{
		"matrix":     state,
		"queue_size": queueSize,
		"history":    history,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) SortHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.Engine.SortMatrixAlphabetically()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "sorted"})
}
