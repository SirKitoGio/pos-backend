package engine

import (
	"log"
	"pos-backend/internal/models"
	"pos-backend/internal/repository"
	"sync"
	"time"
)

// Engine orchestrates the data structures and background processing
type Engine struct {
	Matrix   *Matrix
	BST      *BST
	Stack    *Stack
	AuditLog []models.Action
	mu       sync.RWMutex
	Queue    chan models.Transaction
	Repo     *repository.Repository
}

// NewEngine initializes the core data structures
func NewEngine(repo *repository.Repository) *Engine {
	return &Engine{
		Matrix:   NewMatrix(10, 10), // Example grid size
		BST:      &BST{},
		Stack:    &Stack{},
		AuditLog: []models.Action{},
		Queue:    make(chan models.Transaction, 100),
		Repo:     repo,
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

func (e *Engine) addAuditLog(action models.Action) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.AuditLog = append(e.AuditLog, action)
}

func (e *Engine) GetAuditLog() []models.Action {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	snapshot := make([]models.Action, len(e.AuditLog))
	copy(snapshot, e.AuditLog)
	return snapshot
}

// ClearState resets the in-memory data structures
func (e *Engine) ClearState() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Matrix.Clear()
	e.BST = &BST{}
	e.Stack = &Stack{}
	e.AuditLog = []models.Action{}
}

// processTransaction updates the internal state based on incoming payloads
func (e *Engine) processTransaction(tx models.Transaction) {
	e.mu.Lock()
	defer e.mu.Unlock()
	log.Printf("Processing transaction: %+v", tx)

	var x, y int
	var qty int
	var startTime time.Time

	switch tx.Action {
	case "ADD":
		// 1. Check if item already exists for this specific date
		existing := e.BST.Search(tx.Item, tx.Date)

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
		e.Matrix.Update(x, y, tx.Item, qty, tx.Price, tx.ProductType, tx.Date, startTime, true)

		// 3. Update BST
		e.BST.Insert(models.Item{
			Name:           tx.Item,
			Quantity:       qty,
			Price:          tx.Price,
			ProductType:    tx.ProductType,
			InventoryPlace: tx.InventoryPlace,
			Date:           tx.Date,
			StartTime:      startTime,
			X:              x,
			Y:              y,
		})

		// 4. Push Undo Action to Stack
		e.Stack.Push(models.Action{
			Item:        tx.Item,
			Qty:         tx.Qty,
			Price:       tx.Price,
			ProductType: tx.ProductType,
			Date:        tx.Date,
			Action:      "REMOVE",
			X:           x,
			Y:           y,
			Timestamp:   time.Now(),
		})

	case "REMOVE", "DELETE":
		// Find item coordinates in BST
		existing := e.BST.Search(tx.Item, tx.Date)
		if existing == nil {
			log.Printf("Item %s on date %s not found in BST", tx.Item, tx.Date)
			return
		}

		x, y = existing.X, existing.Y
		newQty := existing.Quantity - tx.Qty
		if tx.Action == "DELETE" || newQty < 0 {
			newQty = 0
		}

		if newQty == 0 {
			// Update Matrix (clear slot)
			e.Matrix.Update(x, y, "", 0, 0, "", "", time.Time{}, false)
			// Delete from BST
			e.BST.Delete(tx.Item, tx.Date)
		} else {
			// Update Matrix with reduced quantity
			e.Matrix.Update(x, y, tx.Item, newQty, existing.Price, existing.ProductType, tx.Date, existing.StartTime, true)
			// Update BST with reduced quantity
			e.BST.Insert(models.Item{
				Name:           tx.Item,
				Quantity:       newQty,
				Price:          existing.Price,
				ProductType:    existing.ProductType,
				InventoryPlace: existing.InventoryPlace,
				Date:           tx.Date,
				StartTime:      existing.StartTime,
				X:              x,
				Y:              y,
			})
		}

		// Push Undo Action to Stack
		e.Stack.Push(models.Action{
			Item:        tx.Item,
			Qty:         tx.Qty,
			Price:       existing.Price,
			ProductType: existing.ProductType,
			Date:        tx.Date,
			Action:      "ADD",
			X:           x,
			Y:           y,
			Timestamp:   time.Now(),
		})

	default:
		log.Printf("Unknown action: %s", tx.Action)
		return
	}

	// 5. Synchronous Log to Persistence Pipeline to prevent 'Split Brain' divergence
	if e.Repo != nil {
		e.Repo.LogTransaction(tx, x, y)
	}
}

// RebuildState reads the transaction log from PostgreSQL and rebuilds the Matrix, BST, and AuditLog
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
		x, y := xs[i], ys[i]
		
		switch tx.Action {
		case "ADD":
			existing := e.BST.Search(tx.Item, tx.Date)
			qty := tx.Qty
			startTime := time.Now()
			if existing != nil {
				qty += existing.Quantity
				startTime = existing.StartTime
			}
			e.Matrix.Update(x, y, tx.Item, qty, tx.Price, tx.ProductType, tx.Date, startTime, true)
			e.BST.Insert(models.Item{
				Name:           tx.Item,
				Quantity:       qty,
				Price:          tx.Price,
				ProductType:    tx.ProductType,
				InventoryPlace: tx.InventoryPlace,
				Date:           tx.Date,
				StartTime:      startTime,
				X:              x,
				Y:              y,
			})
			// Rebuild history stack
			e.Stack.Push(models.Action{
				Item:        tx.Item,
				Qty:         tx.Qty,
				Price:       tx.Price,
				ProductType: tx.ProductType,
				Date:        tx.Date,
				Action:      "REMOVE",
				X:           x,
				Y:           y,
				Timestamp:   time.Now(),
			})
			// Rebuild Audit Log
			e.addAuditLog(models.Action{
				Item:        tx.Item,
				Qty:         tx.Qty,
				Price:       tx.Price,
				ProductType: tx.ProductType,
				Date:        tx.Date,
				Action:      "ADD",
				X:           x,
				Y:           y,
				Timestamp:   time.Now(),
			})

		case "REMOVE", "DELETE":
			existing := e.BST.Search(tx.Item, tx.Date)
			if existing != nil {
				newQty := existing.Quantity - tx.Qty
				if tx.Action == "DELETE" || newQty <= 0 {
					e.Matrix.Update(x, y, "", 0, 0, "", "", time.Time{}, false)
					e.BST.Delete(tx.Item, tx.Date)
				} else {
					e.Matrix.Update(x, y, tx.Item, newQty, existing.Price, existing.ProductType, tx.Date, existing.StartTime, true)
					e.BST.Insert(models.Item{
						Name:           tx.Item,
						Quantity:       newQty,
						Price:          existing.Price,
						ProductType:    existing.ProductType,
						InventoryPlace: existing.InventoryPlace,
						Date:           tx.Date,
						StartTime:      existing.StartTime,
						X:              x,
						Y:              y,
					})
				}
				// Rebuild history stack
				e.Stack.Push(models.Action{
					Item:        tx.Item,
					Qty:         tx.Qty,
					Price:       existing.Price,
					ProductType: existing.ProductType,
					Date:        tx.Date,
					Action:      "ADD",
					X:           x,
					Y:           y,
					Timestamp:   time.Now(),
				})
				// Rebuild Audit Log
				e.addAuditLog(models.Action{
					Item:        tx.Item,
					Qty:         tx.Qty,
					Price:       existing.Price,
					ProductType: existing.ProductType,
					Date:        tx.Date,
					Action:      tx.Action,
					X:           x,
					Y:           y,
					Timestamp:   time.Now(),
				})
			}
		}
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
			e.Matrix.Update(row, col, item.Name, item.Quantity, item.Price, item.ProductType, item.Date, item.StartTime, true)
			// Update BST node with new coordinates
			item.X = row
			item.Y = col
			e.BST.Insert(item)
		}
	}
}

// Undo pops the top action and applies the reverse logic
func (e *Engine) Undo() (*models.Action, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	action, ok := e.Stack.Pop()
	if !ok {
		return nil, false
	}

	log.Printf("Undoing action: %+v", action)

	tx := models.Transaction{
		Item:           action.Item,
		Qty:            action.Qty,
		Price:          action.Price,
		ProductType:    action.ProductType,
		Date:           action.Date,
		Action:         action.Action,
	}

	if action.Action == "ADD" {
		existing := e.BST.Search(action.Item, action.Date)
		newQty := action.Qty
		var startTime time.Time
		inventoryPlace := ""
		if existing != nil {
			newQty += existing.Quantity
			startTime = existing.StartTime
			inventoryPlace = existing.InventoryPlace
		} else {
			startTime = time.Now()
		}
		e.Matrix.Update(action.X, action.Y, action.Item, newQty, action.Price, action.ProductType, action.Date, startTime, true)
		e.BST.Insert(models.Item{
			Name:           action.Item,
			Quantity:       newQty,
			Price:          action.Price,
			ProductType:    action.ProductType,
			InventoryPlace: inventoryPlace,
			Date:           action.Date,
			StartTime:      startTime,
			X:              action.X,
			Y:              action.Y,
		})
	} else if action.Action == "REMOVE" {
		existing := e.BST.Search(action.Item, action.Date)
		if existing != nil {
			newQty := existing.Quantity - action.Qty
			if newQty <= 0 {
				e.Matrix.Update(action.X, action.Y, "", 0, 0, "", "", time.Time{}, false)
				e.BST.Delete(action.Item, action.Date)
			} else {
				e.Matrix.Update(action.X, action.Y, action.Item, newQty, existing.Price, existing.ProductType, action.Date, existing.StartTime, true)
				e.BST.Insert(models.Item{
					Name:           action.Item,
					Quantity:       newQty,
					Price:          existing.Price,
					ProductType:    existing.ProductType,
					InventoryPlace: existing.InventoryPlace,
					Date:           action.Date,
					StartTime:      existing.StartTime,
					X:              action.X,
					Y:              action.Y,
				})
			}
		}
	}

	if e.Repo != nil {
		e.Repo.LogTransaction(tx, action.X, action.Y)
	}

	return &action, true
}
