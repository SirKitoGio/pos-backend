package models

import "time"

type Item struct {
	Name           string    `json:"name"`
	Quantity       int       `json:"quantity"`
	Price          float64   `json:"price"`
	ProductType    string    `json:"product_type"`
	InventoryPlace string    `json:"inventory_place"`
	StartTime      time.Time `json:"start_time"`
	X              int       `json:"x"`
	Y              int       `json:"y"`
}

type Transaction struct {
	Item           string  `json:"item"`
	Qty            int     `json:"qty"`
	Price          float64 `json:"price"`
	ProductType    string  `json:"product_type"`
	InventoryPlace string  `json:"inventory_place"`
	Action         string  `json:"action"` // "ADD" or "REMOVE"
}

type Action struct {
	Item        string    `json:"item"`
	Qty         int       `json:"qty"`
	Price       float64   `json:"price"`
	ProductType string    `json:"product_type"`
	Action      string    `json:"action"` // The reverse of the original action
	X           int       `json:"x"`
	Y           int       `json:"y"`
	Timestamp   time.Time `json:"timestamp"`
}

type Slot struct {
	ItemName    string    `json:"item_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	ProductType string    `json:"product_type"`
	StartTime   time.Time `json:"start_time"`
	IsFull      bool      `json:"is_full"`
}
