# Vento Backend - Project Information Document

This document provides a comprehensive overview of the Vento Backend architecture, explaining the role and functionality of each file and component.

## Core Objective
Vento is a high-performance, in-memory inventory management engine built with Go. It handles high-concurrency requests from POS terminals and visualizes warehouse logistics in real-time using specialized data structures.

---

## Root Directory

### `main.go` (cmd/server/main.go)
The **entry point** of the application.
- **Initialization:** Sets up the database connection string (PostgreSQL).
- **Engine Setup:** Initializes the core `Engine` and repository.
- **Boot Sequence:** Calls `e.RebuildState()` to load previous data from the database into RAM.
- **Background Worker:** Starts the non-blocking ingestion queue processor.
- **API Server:** Sets up HTTP routes and starts the web server on port `8080`.

### `go.mod` & `go.sum`
Go's dependency management files. They define the module name and the specific versions of libraries (like the PostgreSQL driver) required by the project.

### `index.html`
A built-in dashboard for real-time visualization. It allows users to interact with the engine via a web interface to search items, see the warehouse grid, and perform undo operations.

### `PRD.md` & `README.md`
- **PRD (Product Requirement Document):** Outlines the technical specifications, goals, and architectural requirements.
- **README:** Provides installation, setup, and execution instructions.

---

## `internal/` Directory
The `internal` folder contains private code that cannot be imported by other projects. This is where the core logic lives.

### `api/handlers.go`
Contains the **HTTP Request Handlers**.
- **`IngestHandler`:** Receives "ADD/REMOVE" requests and pushes them into the asynchronous Queue.
- **`SearchHandler`:** Uses the BST to find items by name or prefix.
- **`UndoHandler`:** Triggers the rollback of the last action using the Stack.
- **`StateHandler`:** Returns the current view of the 2D Matrix for the frontend grid.
- **`SortHandler`:** Re-organizes the warehouse grid alphabetically.

### `engine/` (Core Logic)
This is the "brain" of the system.
- **`engine.go`:** The orchestrator. It manages the flow between the Queue, Matrix, BST, and Stack. It contains the background worker that processes transactions one by one to ensure thread safety (using Mutexes).
- **`bst.go`:** Implements the **Binary Search Tree**. It ensures that item searches are extremely fast (O(log n)) and keeps items alphabetically sorted.
- **`matrix.go`:** Implements the **2D Warehouse Grid**. It tracks the physical location (X, Y) of every item in the warehouse.
- **`stack.go`:** Implements the **Action Stack (LIFO)**. Every time an action is processed, its "inverse" is saved here so that users can "Undo" mistakes.

### `models/models.go`
Defines the **Data Structures** used throughout the app:
- `Item`: Represents a product in the warehouse.
- `Transaction`: Represents an incoming request (e.g., adding 50 burgers).
- `Action`: Represents a historical change used for the Undo stack.
- `Slot`: Represents a single cell in the warehouse grid.

### `repository/postgres.go`
Handles **Persistence**.
- Even though the engine runs in RAM (fast), it logs every change to a PostgreSQL database.
- If the server restarts, this file provides the data to rebuild the entire state from scratch so no data is lost.

---

## Testing
The project includes unit tests to ensure the reliability of the core data structures.
- **`internal/engine/bst_test.go`:** Verifies BST insertion, deletion, and search logic.
- **`internal/engine/matrix_test.go`:** Ensures the warehouse grid updates correctly and handles slot allocation.
You can run all tests using the command: `go test ./internal/engine/...`

---

## Data Pipeline Flow
1. **Request:** POS sends a POST to `/api/ingest`.
2. **Queue:** The API drops the request into a Go Channel (Queue) and responds `200 OK` immediately.
3. **Worker:** A background worker picks up the request.
4. **Logic:** 
   - Locks the system (Mutex).
   - Updates the **Matrix** (Visual location).
   - Updates the **BST** (Search index).
   - Pushes to the **Stack** (Undo history).
   - Unlocks the system.
5. **Persistence:** Asynchronously writes the change to **PostgreSQL**.
