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

func getCompositeKey(name, date string) string {
	return strings.ToLower(name) + "|" + date
}

func insert(node *Node, item models.Item) *Node {
	if node == nil {
		return &Node{Item: item}
	}

	itemKey := getCompositeKey(item.Name, item.Date)
	nodeKey := getCompositeKey(node.Item.Name, node.Item.Date)

	if itemKey < nodeKey {
		node.Left = insert(node.Left, item)
	} else if itemKey > nodeKey {
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

// Search retrieves an item by name and date
func (t *BST) Search(name, date string) *models.Item {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return search(t.root, getCompositeKey(name, date))
}

func search(node *Node, queryKey string) *models.Item {
	if node == nil {
		return nil
	}

	nodeKey := getCompositeKey(node.Item.Name, node.Item.Date)
	if nodeKey == queryKey {
		return &node.Item
	}

	if queryKey < nodeKey {
		return search(node.Left, queryKey)
	}
	return search(node.Right, queryKey)
}

// SearchPrefix returns all items whose name starts with the given query (case-insensitive)
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

	searchPrefix(node.Left, query, results)

	if strings.HasPrefix(nodeNameLower, query) {
		*results = append(*results, node.Item)
	}

	searchPrefix(node.Right, query, results)
}

// Delete removes an item from the tree by name and date
func (t *BST) Delete(name, date string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.root = deleteNode(t.root, name, date)
}

func deleteNode(node *Node, name, date string) *Node {
	if node == nil {
		return nil
	}

	targetKey := getCompositeKey(name, date)
	nodeKey := getCompositeKey(node.Item.Name, node.Item.Date)

	if targetKey < nodeKey {
		node.Left = deleteNode(node.Left, name, date)
	} else if targetKey > nodeKey {
		node.Right = deleteNode(node.Right, name, date)
	} else {
		// Found it
		if node.Left == nil {
			return node.Right
		} else if node.Right == nil {
			return node.Left
		}

		// Node with two children: Get the inorder successor
		minNode := findMin(node.Right)
		node.Item = minNode.Item
		node.Right = deleteNode(node.Right, minNode.Item.Name, minNode.Item.Date)
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
