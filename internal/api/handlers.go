package api

import (
	"encoding/json"
	"log"
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

	// Fetch history directly from Supabase repository if available
	var history []models.Action
	var err error
	if s.Engine.Repo != nil {
		history, err = s.Engine.Repo.GetHistoryLog()
		if err != nil {
			log.Printf("Error fetching history from Supabase: %v", err)
			history = []models.Action{} // Fallback to empty list
		}
	} else {
		// Fallback to in-memory history if repo is not available
		history = s.Engine.GetAuditLog()
	}

	response := map[string]interface{}{
		"matrix":     state,
		"queue_size": queueSize,
		"history":    history,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding state response: %v", err)
	}
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

func (s *Server) ClearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.Engine.Repo != nil {
		if err := s.Engine.Repo.ClearAllTransactions(); err != nil {
			http.Error(w, "Failed to clear database", http.StatusInternalServerError)
			return
		}
	}

	// Reset in-memory state
	s.Engine.ClearState()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Item string `json:"item"`
		Date string `json:"date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Soft delete by queueing a REMOVE transaction for the full quantity
	existing := s.Engine.BST.Search(req.Item, req.Date)
	if existing == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	tx := models.Transaction{
		Item:           existing.Name,
		Qty:            existing.Quantity,
		Price:          existing.Price,
		ProductType:    existing.ProductType,
		InventoryPlace: existing.InventoryPlace,
		Date:           existing.Date,
		Action:         "DELETE", // Explicitly mark as deleted
	}

	s.Engine.Queue <- tx

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "soft_deleted"})
}
