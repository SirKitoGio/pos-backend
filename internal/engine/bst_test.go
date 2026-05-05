package engine

import (
	"pos-backend/internal/models"
	"testing"
)

func TestBST(t *testing.T) {
	bst := &BST{}
	date := "2024-01-01"

	items := []models.Item{
		{Name: "Burger", Quantity: 10, X: 0, Y: 0, Date: date},
		{Name: "Apple", Quantity: 5, X: 0, Y: 1, Date: date},
		{Name: "Cherry", Quantity: 20, X: 0, Y: 2, Date: date},
	}

	for _, item := range items {
		bst.Insert(item)
	}

	// Test Search
	found := bst.Search("Burger", date)
	if found == nil || found.Name != "Burger" {
		t.Errorf("Expected to find Burger, got %v", found)
	}

	// Test Case Insensitive Search
	found = bst.Search("apple", date)
	if found == nil || found.Name != "Apple" {
		t.Errorf("Expected to find Apple, got %v", found)
	}

	found = bst.Search("Donut", date)
	if found != nil {
		t.Errorf("Expected not to find Donut, got %v", found)
	}

	bst.Insert(models.Item{Name: "Burger", Quantity: 50, X: 1, Y: 1, Date: date})
	found = bst.Search("Burger", date)
	if found.Quantity != 50 {
		t.Errorf("Expected quantity 50, got %d", found.Quantity)
	}

	bst.Delete("Apple", date)
	found = bst.Search("Apple", date)
	if found != nil {
		t.Errorf("Expected Apple to be deleted")
	}

	bst.Insert(models.Item{Name: "Apple", Quantity: 5, X: 0, Y: 1, Date: date})
	bst.Insert(models.Item{Name: "Apricot", Quantity: 3, X: 1, Y: 0, Date: date})
	
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
