package engine

import "testing"

func TestMatrix(t *testing.T) {
	m := NewMatrix(5, 5)

	// Initial find empty should be (0, 0)
	x, y, ok := m.FindFirstEmpty()
	if !ok || x != 0 || y != 0 {
		t.Errorf("Expected first empty to be (0, 0), got (%d, %d)", x, y)
	}

	// Fill (0, 0)
	err := m.Update(0, 0, "Burger", 10, true)
	if err != nil {
		t.Errorf("Error updating matrix: %v", err)
	}

	// Now first empty should be (0, 1)
	x, y, ok = m.FindFirstEmpty()
	if !ok || x != 0 || y != 1 {
		t.Errorf("Expected first empty to be (0, 1), got (%d, %d)", x, y)
	}

	// Test bounds
	err = m.Update(5, 5, "Burger", 10, true)
	if err == nil {
		t.Error("Expected error for out-of-bounds coordinates")
	}

	// Test GetState
	state := m.GetState()
	if state[0][0].ItemName != "Burger" {
		t.Errorf("Expected Burger at (0, 0), got %s", state[0][0].ItemName)
	}
}
