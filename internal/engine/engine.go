package engine

import (
	"log"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
	"time"
)

// Engine orchestrates the data structures and background processing
type Engine struct {
	Matrix *Matrix
	BST    *BST
	Stack  *Stack
	Queue  chan models.Transaction
	Repo   *repository.Repository
}

// NewEngine initializes the core data structures
func NewEngine(repo *repository.Repository) *Engine {
	return &Engine{
		Matrix: NewMatrix(10, 10), // Example grid size
		BST:    &BST{},
		Stack:  &Stack{},
		Queue:  make(chan models.Transaction, 100),
		Repo:   repo,
	}
}

// StartWorker launches the background goroutine to process the queue
func (e *Engine) StartWorker() {
	go func() {
		for tx := range e.Queue {
			e.processTransaction(tx)
		}
	}()
}

// processTransaction updates the internal state based on incoming payloads
func (e *Engine) processTransaction(tx models.Transaction) {
	log.Printf("Processing transaction: %+v", tx)

	switch tx.Action {
	case "ADD":
		// 1. Check if item already exists
		existing := e.BST.Search(tx.Item)
		var x, y int
		var qty int
		var startTime time.Time

		if existing != nil {
			x, y = existing.X, existing.Y
			qty = existing.Quantity + tx.Qty
			startTime = existing.StartTime // Keep original start time
		} else {
			// Find a new slot in the matrix
			var ok bool
			x, y, ok = e.Matrix.FindFirstEmpty()
			if !ok {
				log.Println("Matrix is full, cannot add item")
				return
			}
			qty = tx.Qty
			startTime = time.Now()
		}

		// 2. Update Matrix
		e.Matrix.Update(x, y, tx.Item, qty, tx.Price, startTime, true)

		// 3. Update BST
		e.BST.Insert(models.Item{
			Name:      tx.Item,
			Quantity:  qty,
			Price:     tx.Price,
			StartTime: startTime,
			X:         x,
			Y:         y,
		})

		// 4. Push Undo Action to Stack
		e.Stack.Push(models.Action{
			Item:      tx.Item,
			Qty:       tx.Qty,
			Price:     tx.Price,
			Action:    "REMOVE",
			X:         x,
			Y:         y,
			Timestamp: time.Now(),
		})

		// 5. Async Log to Persistence Pipeline
		if e.Repo != nil {
			go e.Repo.LogTransaction(tx, x, y)
		}

	case "REMOVE":
		// Find item coordinates in BST
		existing := e.BST.Search(tx.Item)
		if existing == nil {
			log.Printf("Item %s not found in BST", tx.Item)
			return
		}

		newQty := existing.Quantity - tx.Qty
		if newQty < 0 {
			newQty = 0
		}

		if newQty == 0 {
			// Update Matrix (clear slot)
			e.Matrix.Update(existing.X, existing.Y, "", 0, 0, time.Time{}, false)
			// Delete from BST
			e.BST.Delete(tx.Item)
		} else {
			// Update Matrix with reduced quantity
			e.Matrix.Update(existing.X, existing.Y, tx.Item, newQty, existing.Price, existing.StartTime, true)
			// Update BST with reduced quantity
			e.BST.Insert(models.Item{
				Name:      tx.Item,
				Quantity:  newQty,
				Price:     existing.Price,
				StartTime: existing.StartTime,
				X:         existing.X,
				Y:         existing.Y,
			})
		}

		// Push Undo Action to Stack
		e.Stack.Push(models.Action{
			Item:      tx.Item,
			Qty:       tx.Qty,
			Price:     existing.Price,
			Action:    "ADD",
			X:         existing.X,
			Y:         existing.Y,
			Timestamp: time.Now(),
		})

		// 5. Async Log to Persistence Pipeline
		if e.Repo != nil {
			go e.Repo.LogTransaction(tx, existing.X, existing.Y)
		}

	default:
		log.Printf("Unknown action: %s", tx.Action)
	}
}

// RebuildState reads the transaction log from PostgreSQL and rebuilds the Matrix and BST
func (e *Engine) RebuildState() error {
	if e.Repo == nil {
		return nil
	}

	txs, xs, ys, err := e.Repo.GetAllTransactions()
	if err != nil {
		return err
	}

	log.Printf("Rebuilding state from %d transactions...", len(txs))

	for i, tx := range txs {
		// Note: xs[i] and ys[i] are stored in the DB, but processTransaction currently finds new slots.
		// For a true rebuild, we might want to respect the stored coordinates.
		// But for now, we'll re-process to maintain consistency with the current logic.
		e.processTransaction(tx)
		_ = xs[i]
		_ = ys[i]
	}

	log.Println("State rebuild complete.")
	return nil
}

// SortMatrixAlphabetically re-arranges the Matrix slots based on alphabetical order of items in BST
func (e *Engine) SortMatrixAlphabetically() {
	// 1. Get all items in alphabetical order from BST
	items := e.BST.GetAllInOrder()

	// 2. Clear Matrix
	e.Matrix.Clear()

	// 3. Re-insert items into Matrix sequentially
	for i, item := range items {
		row := i / e.Matrix.cols
		col := i % e.Matrix.cols

		if row < e.Matrix.rows {
			e.Matrix.Update(row, col, item.Name, item.Quantity, item.Price, item.StartTime, true)
			// Update BST node with new coordinates
			item.X = row
			item.Y = col
			e.BST.Insert(item)
		}
	}
}

// Undo pops the top action and applies the reverse logic
func (e *Engine) Undo() (*models.Action, bool) {
	action, ok := e.Stack.Pop()
	if !ok {
		return nil, false
	}

	log.Printf("Undoing action: %+v", action)

	tx := models.Transaction{
		Item:   action.Item,
		Qty:    action.Qty,
		Price:  action.Price,
		Action: action.Action,
	}

	if action.Action == "ADD" {
		existing := e.BST.Search(action.Item)
		newQty := action.Qty
		var startTime time.Time
		if existing != nil {
			newQty += existing.Quantity
			startTime = existing.StartTime
		} else {
			startTime = time.Now()
		}
		e.Matrix.Update(action.X, action.Y, action.Item, newQty, action.Price, startTime, true)
		e.BST.Insert(models.Item{
			Name:      action.Item,
			Quantity:  newQty,
			Price:     action.Price,
			StartTime: startTime,
			X:         action.X,
			Y:         action.Y,
		})
	} else if action.Action == "REMOVE" {
		existing := e.BST.Search(action.Item)
		if existing != nil {
			newQty := existing.Quantity - action.Qty
			if newQty <= 0 {
				e.Matrix.Update(action.X, action.Y, "", 0, 0, time.Time{}, false)
				e.BST.Delete(action.Item)
			} else {
				e.Matrix.Update(action.X, action.Y, action.Item, newQty, existing.Price, existing.StartTime, true)
				e.BST.Insert(models.Item{
					Name:      action.Item,
					Quantity:  newQty,
					Price:     existing.Price,
					StartTime: existing.StartTime,
					X:         action.X,
					Y:         action.Y,
				})
			}
		}
	}

	if e.Repo != nil {
		go e.Repo.LogTransaction(tx, action.X, action.Y)
	}

	return &action, true
}
