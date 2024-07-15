package pkg

import (
	"errors"
	"fmt"
	"sync"

	"github.com/RoaringBitmap/roaring"
)

type MockStorage[T any] struct {
	nodes        map[uint32]*Node[T]
	dependencies map[uint32]*roaring.Bitmap
	dependents   map[uint32]*roaring.Bitmap
	nameToID     map[string]uint32
	idToName     map[uint32]string
	idCounter    uint32
	fullyCached  bool
	mu           sync.Mutex
}

func NewMockStorage[T any]() *MockStorage[T] {
	return &MockStorage[T]{
		nodes:        make(map[uint32]*Node[T]),
		dependencies: make(map[uint32]*roaring.Bitmap),
		dependents:   make(map[uint32]*roaring.Bitmap),
		nameToID:     make(map[string]uint32),
		idToName:     make(map[uint32]string),
		idCounter:    0,
	}
}

func (m *MockStorage[T]) SaveNode(node *Node[T]) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nameToID[node.Name] = node.Id
	m.idToName[node.Id] = node.Name
	m.nodes[node.Id] = node
	return nil
}

func (m *MockStorage[T]) GetNode(id uint32) (*Node[T], error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	node, exists := m.nodes[id]
	if !exists {
		return nil, errors.New("node not found")
	}
	return node, nil
}

func (m *MockStorage[T]) GetAllKeys() ([]uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]uint32, 0, len(m.nodes))
	for k := range m.nodes {
		keys = append(keys, k)
	}

	return keys, nil
}

func (m *MockStorage[T]) SetDependency(nodeID, neighborID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.nodes[nodeID]; !exists {
		return errors.New("node not found")
	}
	if _, exists := m.nodes[neighborID]; !exists {
		return errors.New("neighbor node not found")
	}
	node := m.nodes[nodeID]
	neighbor := m.nodes[neighborID]
	return node.SetDependency(m, neighbor)
}

func (m *MockStorage[T]) QueryDependents(nodeID uint32) (*roaring.Bitmap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.dependents[nodeID]; !exists {
		return nil, fmt.Errorf("node does not exist")
	}
	return m.nodes[nodeID].QueryDependents(m)
}

func (m *MockStorage[T]) QueryDependencies(nodeID uint32) (*roaring.Bitmap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.dependents[nodeID]; !exists {
		return nil, fmt.Errorf("node does not exist")
	}
	return m.nodes[nodeID].QueryDependencies(m)
}

func (m *MockStorage[T]) GenerateID() (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	return m.idCounter, nil
}

func (m *MockStorage[T]) NameToID(name string) (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.nameToID[name]; !exists {
		return 0, errors.New("node with name not found")
	}
	return m.nameToID[name], nil
}

func (m *MockStorage[T]) IDToName(id uint32) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.idToName[id]; !exists {
		return "", errors.New("node with id not found")
	}
	return m.idToName[id], nil
}