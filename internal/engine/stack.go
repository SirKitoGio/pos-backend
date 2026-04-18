package engine

import (
	"pos-backend/internal/models"
	"sync"
)

// Stack is a thread-safe LIFO storage for actions
type Stack struct {
	mu      sync.Mutex
	actions []models.Action
}

// Push adds a new action to the top of the stack
func (s *Stack) Push(action models.Action) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions = append(s.actions, action)
}

// Pop removes and returns the most recent action
func (s *Stack) Pop() (models.Action, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.actions) == 0 {
		return models.Action{}, false
	}

	lastIdx := len(s.actions) - 1
	action := s.actions[lastIdx]
	s.actions = s.actions[:lastIdx]
	return action, true
}

// Size returns the current number of actions in the stack
func (s *Stack) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.actions)
}

// GetActions returns a copy of the current stack for the UI
func (s *Stack) GetActions() []models.Action {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Return a copy to avoid race conditions
	actions := make([]models.Action, len(s.actions))
	copy(actions, s.actions)
	return actions
}
