package repository

import (
	"encoding/json"
	"log"
	"time"
	"pos-backend/internal/models"

	"github.com/supabase-community/postgrest-go"
)

type Repository struct {
	client *postgrest.Client
}

func NewRepository(url, anonKey string) (*Repository, error) {
	if url == "" || anonKey == "" {
		return nil, log.New(log.Writer(), "repo: ", log.LstdFlags).Output(2, "SUPABASE_URL or SUPABASE_ANON_KEY missing")
	}

	client := postgrest.NewClient(url+"/rest/v1", "", map[string]string{
		"apikey":        anonKey,
		"Authorization": "Bearer " + anonKey,
	})

	return &Repository{client: client}, nil
}

// Verify checks if the database is reachable and credentials are valid
func (r *Repository) Verify() error {
	// We attempt to select one row from the transactions table to verify the connection and RLS policies
	resp, _, err := r.client.From("transactions").Select("id", "exact", false).Limit(1, "").Execute()
	if err != nil {
		return err
	}

	// If we get here, the API reached the database. 
	// Note: An empty response is fine as long as err is nil.
	log.Printf("Database connection verified successfully. Response size: %d bytes", len(resp))
	return nil
}

func (r *Repository) InitSchema() error {
	// With PostgREST, we assume the schema is already created via the SQL script
	log.Println("Database schema check via API...")
	return nil
}

func (r *Repository) LogTransaction(tx models.Transaction, x, y int) {
	row := map[string]interface{}{
		"item_name":       tx.Item,
		"quantity":        tx.Qty,
		"price":           tx.Price,
		"product_type":    tx.ProductType,
		"inventory_place": tx.InventoryPlace,
		"date":            tx.Date,
		"action":          tx.Action,
		"x":               x,
		"y":               y,
	}

	body, _, err := r.client.From("transactions").Insert(row, false, "", "", "").Execute()
	if err != nil {
		log.Printf("Failed to log transaction via API. Error: %v | Response Body: %s", err, string(body))
	} else {
		log.Printf("Successfully logged transaction for item: %s", tx.Item)
	}
}

func (r *Repository) GetAllTransactions() ([]models.Transaction, []int, []int, error) {
	result, _, err := r.client.From("transactions").Select("*", "exact", false).Order("created_at", &postgrest.OrderOpts{Ascending: true}).Execute()
	if err != nil {
		return nil, nil, nil, err
	}

	var data []struct {
		ItemName       string  `json:"item_name"`
		Quantity       int     `json:"quantity"`
		Price          float64 `json:"price"`
		ProductType    string  `json:"product_type"`
		InventoryPlace string  `json:"inventory_place"`
		Date           string  `json:"date"`
		Action         string  `json:"action"`
		X              int     `json:"x"`
		Y              int     `json:"y"`
	}

	if err := json.Unmarshal(result, &data); err != nil {
		return nil, nil, nil, err
	}

	var txs []models.Transaction
	var xs []int
	var ys []int

	for _, d := range data {
		txs = append(txs, models.Transaction{
			Item:           d.ItemName,
			Qty:            d.Quantity,
			Price:          d.Price,
			ProductType:    d.ProductType,
			InventoryPlace: d.InventoryPlace,
			Date:           d.Date,
			Action:         d.Action,
		})
		xs = append(xs, d.X)
		ys = append(ys, d.Y)
	}

	return txs, xs, ys, nil
}

func (r *Repository) GetHistoryLog() ([]models.Action, error) {
	// Limit to last 50 transactions for performance on the polled endpoint
	result, _, err := r.client.From("transactions").Select("*", "exact", false).Order("created_at", &postgrest.OrderOpts{Ascending: false}).Limit(50, "").Execute()
	if err != nil {
		return nil, err
	}

	var data []struct {
		ItemName       string  `json:"item_name"`
		Quantity       int     `json:"quantity"`
		Price          float64 `json:"price"`
		ProductType    string  `json:"product_type"`
		Date           string  `json:"date"`
		Action         string  `json:"action"`
		X              int     `json:"x"`
		Y              int     `json:"y"`
		CreatedAt      string  `json:"created_at"`
	}

	if err := json.Unmarshal(result, &data); err != nil {
		return nil, err
	}

	var history []models.Action
	for _, d := range data {
		// Parse with RFC3339Nano to handle Supabase fractional seconds
		t, parseErr := time.Parse(time.RFC3339Nano, d.CreatedAt)
		if parseErr != nil {
			// Fallback to RFC3339 if Nano fails
			t, _ = time.Parse(time.RFC3339, d.CreatedAt)
		}
		
		history = append(history, models.Action{
			Item:        d.ItemName,
			Qty:         d.Quantity,
			Price:       d.Price,
			ProductType: d.ProductType,
			Date:        d.Date,
			Action:      d.Action,
			X:           d.X,
			Y:           d.Y,
			Timestamp:   t,
		})
	}
	
	// We return it in ascending order (oldest to newest) as expected by the frontend logic
	// But we fetched descending to get the "latest 50"
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
	
	return history, nil
}

func (r *Repository) ClearAllTransactions() error {
	_, _, err := r.client.From("transactions").Delete("", "").Not("id", "is", "null").Execute()
	if err != nil {
		log.Printf("Failed to clear transactions via API. Error: %v", err)
		return err
	}
	log.Println("Successfully cleared all transactions from database.")
	return nil
}
