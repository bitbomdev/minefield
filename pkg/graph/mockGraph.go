package graph

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/RoaringBitmap/roaring"
)

type MockStorage struct {
	nodes        map[uint32]*Node
	dependencies map[uint32]*roaring.Bitmap
	dependents   map[uint32]*roaring.Bitmap
	nameToID     map[string]uint32
	cache        map[uint32]*NodeCache
	toBeCached   []uint32
	mu           sync.Mutex
	idCounter    uint32
	fullyCached  bool
	db           map[string]map[string][]byte

	// Error injection fields
	SaveNodeErr              error
	GetNodeErr               error
	GetNodesByGlobErr        error
	GetAllKeysErr            error
	SaveCacheErr             error
	ToBeCachedErr            error
	AddNodeToCachedStackErr  error
	ClearCacheStackErr       error
	GetCacheErr              error
	GenerateIDErr            error
	NameToIDErr              error
	GetNodesErr              error
	SaveCachesErr            error
	GetCachesErr             error
	RemoveAllCachesErr       error
	AddOrUpdateCustomDataErr error
	GetCustomDataErr         error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		nodes:        make(map[uint32]*Node),
		dependencies: make(map[uint32]*roaring.Bitmap),
		dependents:   make(map[uint32]*roaring.Bitmap),
		nameToID:     make(map[string]uint32),
		idCounter:    0,
		db:           make(map[string]map[string][]byte),
	}
}

func (m *MockStorage) SaveNode(node *Node) error {
	if m.SaveNodeErr != nil {
		return m.SaveNodeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nameToID[node.Name] = node.ID
	m.nodes[node.ID] = node
	m.toBeCached = append(m.toBeCached, node.ID)
	return nil
}

func (m *MockStorage) GetNode(id uint32) (*Node, error) {
	if m.GetNodeErr != nil {
		return nil, m.GetNodeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	node, exists := m.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node %v not found", id)
	}
	return node, nil
}

func (m *MockStorage) GetNodesByGlob(pattern string) ([]*Node, error) {
	if m.GetNodesByGlobErr != nil {
		return nil, m.GetNodesByGlobErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	nodes := make([]*Node, 0)
	for _, node := range m.nodes {
		matched, err := filepath.Match(pattern, node.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern: %v", err)
		}
		if matched {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func (m *MockStorage) GetAllKeys() ([]uint32, error) {
	if m.GetAllKeysErr != nil {
		return nil, m.GetAllKeysErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]uint32, 0, len(m.nodes))
	for k := range m.nodes {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *MockStorage) SaveCache(cache *NodeCache) error {
	if m.SaveCacheErr != nil {
		return m.SaveCacheErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cache == nil {
		m.cache = map[uint32]*NodeCache{}
	}
	m.cache[cache.ID] = cache
	return nil
}

func (m *MockStorage) ToBeCached() ([]uint32, error) {
	if m.ToBeCachedErr != nil {
		return nil, m.ToBeCachedErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.toBeCached, nil
}

func (m *MockStorage) AddNodeToCachedStack(id uint32) error {
	if m.AddNodeToCachedStackErr != nil {
		return m.AddNodeToCachedStackErr
	}
	m.toBeCached = append(m.toBeCached, id)
	return nil
}

func (m *MockStorage) ClearCacheStack() error {
	if m.ClearCacheStackErr != nil {
		return m.ClearCacheStackErr
	}
	m.toBeCached = []uint32{}
	return nil
}

func (m *MockStorage) GetCache(id uint32) (*NodeCache, error) {
	if m.GetCacheErr != nil {
		return nil, m.GetCacheErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.cache[id]; !ok {
		return nil, errors.New("cache not found")
	}
	return m.cache[id], nil
}

func (m *MockStorage) GenerateID() (uint32, error) {
	if m.GenerateIDErr != nil {
		return 0, m.GenerateIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return m.idCounter, nil
}

func (m *MockStorage) NameToID(name string) (uint32, error) {
	if m.NameToIDErr != nil {
		return 0, m.NameToIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if id, exists := m.nameToID[name]; exists {
		return id, nil
	}
	return 0, errors.New("node with name not found")
}

func (m *MockStorage) GetNodes(ids []uint32) (map[uint32]*Node, error) {
	if m.GetNodesErr != nil {
		return nil, m.GetNodesErr
	}
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

func (m *MockStorage) SaveCaches(caches []*NodeCache) error {
	if m.SaveCachesErr != nil {
		return m.SaveCachesErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cache == nil {
		m.cache = map[uint32]*NodeCache{}
	}
	for _, cache := range caches {
		m.cache[cache.ID] = cache
	}
	return nil
}

func (m *MockStorage) GetCaches(ids []uint32) (map[uint32]*NodeCache, error) {
	if m.GetCachesErr != nil {
		return nil, m.GetCachesErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	caches := make(map[uint32]*NodeCache, len(ids))
	for _, id := range ids {
		cache, exists := m.cache[id]
		if !exists {
			continue // Skip missing caches
		}
		caches[id] = cache
	}

	return caches, nil
}

func (m *MockStorage) RemoveAllCaches() error {
	if m.RemoveAllCachesErr != nil {
		return m.RemoveAllCachesErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add all cache IDs to the toBeCached slice
	for id := range m.cache {
		m.toBeCached = append(m.toBeCached, id)
	}

	// Clear the cache
	m.cache = make(map[uint32]*NodeCache)

	return nil
}

func (m *MockStorage) AddOrUpdateCustomData(tag, key, dataKey string, data []byte) error {
	if m.AddOrUpdateCustomDataErr != nil {
		return m.AddOrUpdateCustomDataErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	fullKey := fmt.Sprintf("%s:%s", tag, key)
	if m.db[fullKey] == nil {
		m.db[fullKey] = make(map[string][]byte)
	}
	m.db[fullKey][dataKey] = data
	return nil
}

func (m *MockStorage) GetCustomData(tag, key string) (map[string][]byte, error) {
	if m.GetCustomDataErr != nil {
		return nil, m.GetCustomDataErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	fullKey := fmt.Sprintf("%s:%s", tag, key)
	data, exists := m.db[fullKey]
	if !exists {
		return nil, fmt.Errorf("no data found for tag: %s, key: %s", tag, key)
	}
	return data, nil
}
