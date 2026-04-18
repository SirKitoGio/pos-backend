package engine

import (
	"pos-backend/internal/models"
	"strings"
	"sync"
)

type BST struct {
	mu   sync.RWMutex
	root *Node
}

type Node struct {
	Item  models.Item
	Left  *Node
	Right *Node
}

func (t *BST) Insert(item models.Item) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = insert(t.root, item)
}

func insert(node *Node, item models.Item) *Node {
	if node == nil {
		return &Node{Item: item}
	}

	if item.Name < node.Item.Name {
		node.Left = insert(node.Left, item)
	} else if item.Name > node.Item.Name {
		node.Right = insert(node.Right, item)
	} else {
		// Update existing item
		node.Item.Quantity = item.Quantity
		node.Item.Price = item.Price
		node.Item.StartTime = item.StartTime
		node.Item.X = item.X
		node.Item.Y = item.Y
	}
	return node
}

// Search retrieves an item by name (exact or partial)
func (t *BST) Search(query string) *models.Item {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return search(t.root, strings.ToLower(query))
}

func search(node *Node, query string) *models.Item {
	if node == nil {
		return nil
	}

	nodeNameLower := strings.ToLower(node.Item.Name)
	if nodeNameLower == query {
		return &node.Item
	}

	if query < nodeNameLower {
		return search(node.Left, query)
	}
	return search(node.Right, query)
}

// SearchPrefix returns all items starting with the given query (case-insensitive)
func (t *BST) SearchPrefix(query string) []models.Item {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var results []models.Item
	query = strings.ToLower(query)
	searchPrefix(t.root, query, &results)
	return results
}

func searchPrefix(node *Node, query string, results *[]models.Item) {
	if node == nil {
		return
	}

	nodeNameLower := strings.ToLower(node.Item.Name)

	// Visit left subtree if query could match items there
	// We always visit both unless we're sure the query can't match.
	// For a prefix search, a node's name being greater than the query
	// doesn't mean its left child can't match.
	// But if query is greater than the node's prefix, we might skip some.
	// For simplicity and correctness in a small tree, a full traversal is safe,
	// but we can optimize.

	searchPrefix(node.Left, query, results)

	if strings.HasPrefix(nodeNameLower, query) {
		*results = append(*results, node.Item)
	}

	searchPrefix(node.Right, query, results)
}

// Delete removes an item from the tree
func (t *BST) Delete(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = delete(t.root, name)
}

func delete(node *Node, name string) *Node {
	if node == nil {
		return nil
	}

	if name < node.Item.Name {
		node.Left = delete(node.Left, name)
	} else if name > node.Item.Name {
		node.Right = delete(node.Right, name)
	} else {
		// Found it
		if node.Left == nil {
			return node.Right
		} else if node.Right == nil {
			return node.Left
		}

		// Node with two children: Get the inorder successor (smallest in the right subtree)
		minNode := findMin(node.Right)
		node.Item = minNode.Item
		node.Right = delete(node.Right, minNode.Item.Name)
	}
	return node
}

// GetAllInOrder returns all items in alphabetical order using in-order traversal
func (t *BST) GetAllInOrder() []models.Item {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var items []models.Item
	getAllInOrder(t.root, &items)
	return items
}

func getAllInOrder(node *Node, items *[]models.Item) {
	if node == nil {
		return
	}
	getAllInOrder(node.Left, items)
	*items = append(*items, node.Item)
	getAllInOrder(node.Right, items)
}

func findMin(node *Node) *Node {
	current := node
	for current.Left != nil {
		current = current.Left
	}
	return current
}
