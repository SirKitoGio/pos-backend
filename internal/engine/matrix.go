package engine

import (
	"fmt"
	"pos-backend/internal/models"
	"sync"
	"time"
)

type Matrix struct {
	mu    sync.RWMutex
	grid  [][]models.Slot
	rows  int
	cols  int
}

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

func (m *Matrix) Update(x, y int, item string, qty int, price float64, productType string, startTime time.Time, isFull bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if x < 0 || x >= m.rows || y < 0 || y >= m.cols {
		return fmt.Errorf("coordinates (%d, %d) out of bounds", x, y)
	}

	m.grid[x][y] = models.Slot{
		ItemName:    item,
		Quantity:    qty,
		Price:       price,
		ProductType: productType,
		StartTime:   startTime,
		IsFull:      isFull,
	}
	return nil
}

func (m *Matrix) GetState() [][]models.Slot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make([][]models.Slot, m.rows)
	for i := range m.grid {
		snapshot[i] = make([]models.Slot, m.cols)
		copy(snapshot[i], m.grid[i])
	}
	return snapshot
}

func (m *Matrix) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.grid {
		for j := range m.grid[i] {
			m.grid[i][j] = models.Slot{IsFull: false}
		}
	}
}

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
