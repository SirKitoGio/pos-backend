package repository

import (
	"database/sql"
	"log"
	"pos-backend/internal/models"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(connStr string) (*Repository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	repo := &Repository{db: db}
	if err := repo.InitSchema(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) InitSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			item_name TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			price DOUBLE PRECISION NOT NULL DEFAULT 0,
			product_type TEXT NOT NULL DEFAULT '',
			inventory_place TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`ALTER TABLE transactions ADD COLUMN IF NOT EXISTS price DOUBLE PRECISION NOT NULL DEFAULT 0;`,
		`ALTER TABLE transactions ADD COLUMN IF NOT EXISTS product_type TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE transactions ADD COLUMN IF NOT EXISTS inventory_place TEXT NOT NULL DEFAULT '';`,
	}

	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) LogTransaction(tx models.Transaction, x, y int) {
	query := `INSERT INTO transactions (item_name, quantity, price, product_type, inventory_place, action, x, y) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query, tx.Item, tx.Qty, tx.Price, tx.ProductType, tx.InventoryPlace, tx.Action, x, y)
	if err != nil {
		log.Printf("Failed to log transaction to DB: %v", err)
	}
}

func (r *Repository) GetAllTransactions() ([]models.Transaction, []int, []int, error) {
	query := `SELECT item_name, quantity, price, product_type, inventory_place, action, x, y FROM transactions ORDER BY created_at ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	var xs []int
	var ys []int

	for rows.Next() {
		var tx models.Transaction
		var x, y int
		if err := rows.Scan(&tx.Item, &tx.Qty, &tx.Price, &tx.ProductType, &tx.InventoryPlace, &tx.Action, &x, &y); err != nil {
			return nil, nil, nil, err
		}
		txs = append(txs, tx)
		xs = append(xs, x)
		ys = append(ys, y)
	}

	return txs, xs, ys, nil
}
