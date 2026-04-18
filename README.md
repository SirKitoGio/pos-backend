# POS Inventory Backend Engine

## Project Overview
This project is a high-performance, in-memory inventory management engine built with Go. It is designed to handle high-concurrency requests from POS terminals and visualize warehouse logistics in real-time.

The system utilizes several core data structures to manage state:
- Binary Search Tree (BST): Used for efficient, alphabetical item searching.
- 2D Matrix: Represents the physical layout of the warehouse and tracks item coordinates.
- Action Stack: Implements a Last-In-First-Out (LIFO) mechanism for undoing operations.
- Ingestion Queue: A buffered channel that processes incoming transactions asynchronously to ensure the main thread remains non-blocking.

A background worker process consumes the ingestion queue, updates the in-memory state with mutex locking for thread safety, and persists transactions to a PostgreSQL database for durability.

## Dependencies
- Go 1.26.1 or higher
- PostgreSQL (Optional for persistence)
- lib/pq (Go PostgreSQL driver)

## Installation and Setup
To set up the project locally, follow these steps:

1. Clone the repository:
   git clone <repository-url>
   cd pos-backend

2. Install dependencies:
   go mod download

3. (Optional) PostgreSQL Setup:
   If you have PostgreSQL installed and want to enable persistence, create a database:
   createdb pos

   Then set the environment variable:
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/pos?sslmode=disable"

## Running the Application
To start the backend server and the inventory engine, run:

go run cmd/server/main.go

The server will start listening on port 8080.

## Running Tests
To verify the engine logic and data structures:

go test ./internal/engine/...

## Accessing the Dashboard
Once the server is running, you can access the built-in desktop dashboard by navigating to:
http://localhost:8080

This dashboard allows you to search the inventory, ingest new items with price and quantity, visualize the warehouse grid, and perform undo operations.
