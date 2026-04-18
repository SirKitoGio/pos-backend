Backend PRD: Go (Golang) Inventory Engine & API

1. Objective
To build a highly concurrent, in-memory data engine using Go. The backend must receive HTTP requests from the Flutter POS terminals, queue them asynchronously, update the core data structures (Matrix, BST, Stack) without blocking the main thread, and serve the updated state back to the UI in real-time.

2. Core Data Structures (In-Memory State)
The entire state of the warehouse lives in the server's RAM. To prevent race conditions (e.g., two cashiers trying to add items to the exact same Matrix slot at the same millisecond), the engine will use Mutexes (Mutual Exclusion locks) to secure the data.
The Matrix (2D Array): * Implemented as a 2D slice ([][]Slot).
Tracks the exact x and y coordinates of every item.
The Binary Search Tree (BST): * A custom Node struct containing the item's string name, total quantity, and a pointer to its Matrix coordinates.
Maintains alphabetical sorting for instant search.
The Ingestion Queue: * A buffered Go Channel (chan Transaction) or a custom Linked List.
Holds incoming JSON payloads before they are written to the Matrix.
The Action Stack: * A LIFO (Last-In-First-Out) array of Action structs.
Stores the exact inverse operation of the last processed task.
3. API Endpoints (The REST Bridge)
The Flutter frontend will communicate with this engine using standard JSON over HTTP.
POST /api/ingest
Payload: { "item": "Burgers", "qty": 500, "action": "ADD" }
Behavior: Instantly pushes the payload into the Queue and returns a 200 OK to the frontend. It does not wait for the Matrix to update.
GET /api/search?q={query}
Behavior: Traverses the BST looking for partial or exact matches. Returns the item data and its Matrix coordinates.
POST /api/undo
Behavior: Pops the top action off the Stack, applies the reverse logic to the Matrix/BST, and returns the updated grid coordinates.
GET /api/state
Behavior: Returns the current visual state of the 2D Matrix and the pending Queue so the Flutter GridView can render the warehouse accurately.
4. Concurrency & Pipeline Logic (The Go Worker)
This is where treating the system like a localized data pipeline shines. The backend operates in two distinct, concurrent layers:
The HTTP Layer (Main Thread): Listens for incoming requests from the POS. When a cashier scans an item, this layer simply drops the data into the Queue and immediately goes back to listening.
The Engine Layer (Background Goroutine): A dedicated worker function that runs in an infinite loop.
It listens to the Queue.
When an item appears, it locks the Mutex.
It updates the Matrix, updates the BST, and pushes the rollback command to the Stack.
It unlocks the Mutex, ready for the next item.
5. Persistence & Safety
Because RAM is wiped if the EC2 instance restarts, the background worker needs a persistence pipeline.
Write-Behind Logging: Every time the background Goroutine successfully updates the Matrix and unlocks the Mutex, it fires off a lightweight, asynchronous SQL query to a PostgreSQL database (e.g., INSERT INTO transaction_log...).
Boot Sequence: If the server ever crashes, the main.go initialization function reads that PostgreSQL log and rapidly rebuilds the Matrix, BST, and Stack in memory before opening the API to the cashiers.

