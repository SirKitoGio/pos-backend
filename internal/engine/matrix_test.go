package engine

import (
	"testing"
	"time"
)

func TestMatrix(t *testing.T) {
	m := NewMatrix(5, 5)

	x, y, ok := m.FindFirstEmpty()
	if !ok || x != 0 || y != 0 {
		t.Errorf("Expected first empty to be (0, 0), got (%d, %d)", x, y)
	}

	err := m.Update(0, 0, "Burger", 10, 5.99, "Food", "2024-01-01", time.Now(), true)
	if err != nil {
		t.Errorf("Error updating matrix: %v", err)
	}

	x, y, ok = m.FindFirstEmpty()
	if !ok || x != 0 || y != 1 {
		t.Errorf("Expected first empty to be (0, 1), got (%d, %d)", x, y)
	}

	err = m.Update(5, 5, "Burger", 10, 5.99, "Food", "2024-01-01", time.Now(), true)
	if err == nil {
		t.Error("Expected error for out-of-bounds coordinates")
	}

	state := m.GetState()
	if state[0][0].ItemName != "Burger" {
		t.Errorf("Expected Burger at (0, 0), got %s", state[0][0].ItemName)
	}
}
