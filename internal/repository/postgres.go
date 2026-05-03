package repository

import (
	"encoding/json"
	"log"
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
			Action:         d.Action,
		})
		xs = append(xs, d.X)
		ys = append(ys, d.Y)
	}

	return txs, xs, ys, nil
}
