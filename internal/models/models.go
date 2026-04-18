package models

import "time"

// Item represents the core data structure stored in the BST
type Item struct {
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	StartTime time.Time `json:"start_time"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
}

// Transaction represents an incoming payload from the Flutter POS
type Transaction struct {
	Item   string  `json:"item"`
	Qty    int     `json:"qty"`
	Price  float64 `json:"price"`
	Action string  `json:"action"` // "ADD" or "REMOVE"
}

// Action represents the inverse operation for the Action Stack (Undo)
type Action struct {
	Item      string    `json:"item"`
	Qty       int       `json:"qty"`
	Price     float64   `json:"price"`
	Action    string    `json:"action"` // The reverse of the original action
	X         int       `json:"x"`
	Y         int       `json:"y"`
	Timestamp time.Time `json:"timestamp"`
}

// Slot represents a coordinate in the 2D Matrix
type Slot struct {
	ItemName  string    `json:"item_name"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	StartTime time.Time `json:"start_time"`
	IsFull    bool      `json:"is_full"`
}
