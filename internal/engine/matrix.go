package engine

import (
	"fmt"
	"pos-backend/internal/models"
	"sync"
	"time"
)

// Matrix holds the 2D visual state of the warehouse
type Matrix struct {
	mu   sync.RWMutex
	grid [][]models.Slot
	rows int
	cols int
}

// NewMatrix initializes a new fixed-size grid
func NewMatrix(rows, cols int) *Matrix {
	grid := make([][]models.Slot, rows)
	for i := range grid {
		grid[i] = make([]models.Slot, cols)
	}
	return &Matrix{
		grid: grid,
		rows: rows,
		cols: cols,
	}
}

// Update updates an item at the specified coordinates
func (m *Matrix) Update(x, y int, item string, qty int, price float64, startTime time.Time, isFull bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if x < 0 || x >= m.rows || y < 0 || y >= m.cols {
		return fmt.Errorf("coordinates (%d, %d) out of bounds", x, y)
	}

	m.grid[x][y] = models.Slot{
		ItemName:  item,
		Quantity:  qty,
		Price:     price,
		StartTime: startTime,
		IsFull:    isFull,
	}
	return nil
}

// GetState returns a snapshot of the grid for the API
func (m *Matrix) GetState() [][]models.Slot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a deep copy to avoid race conditions when the API reads it
	snapshot := make([][]models.Slot, m.rows)
	for i := range m.grid {
		snapshot[i] = make([]models.Slot, m.cols)
		copy(snapshot[i], m.grid[i])
	}
	return snapshot
}

// Clear resets all slots in the matrix
func (m *Matrix) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.grid {
		for j := range m.grid[i] {
			m.grid[i][j] = models.Slot{IsFull: false}
		}
	}
}

// FindFirstEmpty finds the first available slot in the matrix
func (m *Matrix) FindFirstEmpty() (int, int, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			if !m.grid[i][j].IsFull {
				return i, j, true
			}
		}
	}
	return -1, -1, false
}
