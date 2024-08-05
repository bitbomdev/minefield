package pkg

import (
	"errors"
	"fmt"
	"sync"

	"github.com/RoaringBitmap/roaring"
)

type MockStorage struct {
	nodes        map[uint32]*Node
	dependencies map[uint32]*roaring.Bitmap
	dependents   map[uint32]*roaring.Bitmap
	nameToID     map[string]uint32
	idCounter    uint32
	fullyCached  bool
	mu           sync.Mutex
	cache        map[uint32]*NodeCache
	toBeCached   []uint32
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		nodes:        make(map[uint32]*Node),
		dependencies: make(map[uint32]*roaring.Bitmap),
		dependents:   make(map[uint32]*roaring.Bitmap),
		nameToID:     make(map[string]uint32),
		idCounter:    0,
	}
}

func (m *MockStorage) SaveNode(node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nameToID[node.Name] = node.ID
	m.nodes[node.ID] = node
	return nil
}

func (m *MockStorage) GetNode(id uint32) (*Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	node, exists := m.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %v not found", id)
	}
	return node, nil
}

func (m *MockStorage) GetAllKeys() ([]uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]uint32, 0, len(m.nodes))
	for k := range m.nodes {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *MockStorage) SaveCache(cache *NodeCache) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cache == nil {
		m.cache = map[uint32]*NodeCache{}
	}
	m.cache[cache.nodeID] = cache
	if err := m.AddNodeToCachedStack(cache.nodeID); err != nil {
		return err
	}
	return nil
}

func (m *MockStorage) ToBeCached() ([]uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.toBeCached, nil
}

func (m *MockStorage) AddNodeToCachedStack(id uint32) error {
	m.toBeCached = append(m.toBeCached, id)

	return nil
}

func (m *MockStorage) ClearCacheStack() error {
	m.toBeCached = []uint32{}

	return nil
}

func (m *MockStorage) GetCache(id uint32) (*NodeCache, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.cache[id]; !ok {
		return nil, errors.New("cacheHelper not found")
	}
	return m.cache[id], nil
}

func (m *MockStorage) GenerateID() (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return m.idCounter, nil
}

func (m *MockStorage) NameToID(name string) (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.nameToID[name]; !exists {
		return 0, errors.New("node with name not found")
	}
	return m.nameToID[name], nil
}

func (m *MockStorage) GetNodes(ids []uint32) (map[uint32]*Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	nodes := make(map[uint32]*Node, len(ids))
	for _, id := range ids {
		node, exists := m.nodes[id]
		if !exists {
			continue // Skip missing nodes
		}
		nodes[id] = node
	}

	return nodes, nil
}