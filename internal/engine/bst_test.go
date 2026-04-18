package engine

import (
	"pos-backend/internal/models"
	"testing"
	"time"
)

func TestBST(t *testing.T) {
	bst := &BST{}
	now := time.Now()

	items := []models.Item{
		{Name: "Burger", Quantity: 10, Price: 5.99, StartTime: now, X: 0, Y: 0},
		{Name: "Apple", Quantity: 5, Price: 0.99, StartTime: now, X: 0, Y: 1},
		{Name: "Cherry", Quantity: 20, Price: 2.50, StartTime: now, X: 0, Y: 2},
	}

	for _, item := range items {
		bst.Insert(item)
	}

	// Test Search
	found := bst.Search("Burger")
	if found == nil || found.Name != "Burger" {
		t.Errorf("Expected to find Burger, got %v", found)
	}

	// Test Case Insensitive Search
	found = bst.Search("apple")
	if found == nil || found.Name != "Apple" {
		t.Errorf("Expected to find Apple, got %v", found)
	}

	// Test Not Found
	found = bst.Search("Donut")
	if found != nil {
		t.Errorf("Expected not to find Donut, got %v", found)
	}

	// Test Update
	bst.Insert(models.Item{Name: "Burger", Quantity: 50, Price: 5.99, StartTime: now, X: 1, Y: 1})
	found = bst.Search("Burger")
	if found.Quantity != 50 {
		t.Errorf("Expected quantity 50, got %d", found.Quantity)
	}

	// Test Delete
	bst.Delete("Apple")
	found = bst.Search("Apple")
	if found != nil {
		t.Errorf("Expected Apple to be deleted")
	}

	// Test SearchPrefix
	bst.Insert(models.Item{Name: "Apple", Quantity: 5, Price: 0.99, StartTime: now, X: 0, Y: 1})
	bst.Insert(models.Item{Name: "Apricot", Quantity: 3, Price: 1.50, StartTime: now, X: 1, Y: 0})
	
	results := bst.SearchPrefix("Ap")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for prefix 'Ap', got %d", len(results))
	}

	results = bst.SearchPrefix("apple")
	if len(results) != 1 || results[0].Name != "Apple" {
		t.Errorf("Expected 1 result 'Apple' for prefix 'apple', got %v", results)
	}

	results = bst.SearchPrefix("Z")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for prefix 'Z', got %d", len(results))
	}
}
