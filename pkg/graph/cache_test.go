package graph

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_findCycles(t *testing.T) {
	storage := NewMockStorage()
	node1, err := AddNode(storage, "type1", "metadata1", "1")
	assert.NoError(t, err)
	node2, err := AddNode(storage, "type2", "metadata2", "2")
	assert.NoError(t, err)
	err = node1.SetDependency(storage, node2)
	assert.NoError(t, err)

	allNodes, err := storage.GetNodes([]uint32{node1.ID, node2.ID})
	assert.NoError(t, err)

	got := findCycles(2, allNodes)
	assert.Equal(t, map[uint32]uint32{1: 1, 2: 2}, got)
}

func Test_findCycles_With_Cycles(t *testing.T) {
	storage := NewMockStorage()
	node1, err := AddNode(storage, "type1", "metadata1", "1")
	assert.NoError(t, err)
	node2, err := AddNode(storage, "type2", "metadata2", "2")
	assert.NoError(t, err)
	node3, err := AddNode(storage, "type3", "metadata3", "3")
	assert.NoError(t, err)

	err = node1.SetDependency(storage, node2)
	assert.NoError(t, err)
	err = node2.SetDependency(storage, node3)
	assert.NoError(t, err)
	err = node3.SetDependency(storage, node1)
	assert.NoError(t, err)

	allNodes, err := storage.GetNodes([]uint32{node1.ID, node2.ID, node3.ID})
	assert.NoError(t, err)

	got := findCycles(3, allNodes)

	assert.Equal(t, map[uint32]uint32{1: 1, 2: 1, 3: 1}, got)
}

func TestRandomGraphDependenciesWithControlledCircles(t *testing.T) {
	tests := []int{1000}
	for _, n := range tests {
		storage := NewMockStorage()
		nodes := make([]*Node, n)
		expectedDependents := make(map[uint32][]uint32)
		expectedDependencies := make(map[uint32][]uint32)

		// Create nodes and set dependencies

		for i := 0; i < n; i++ {
			node, err := AddNode(storage, fmt.Sprintf("type %d", i+1), fmt.Sprintf("metadata %d", i), fmt.Sprintf("name %d", i+1))
			assert.NoError(t, err)
			nodes[i] = node
		}

		// Set random dependencies, allowing controlled circles
		cycleProbability := 0.01 // 1% chance to create a cycle
		for i := 0; i < n; i++ {
			possibleDeps := rand.Perm(n - i)                   // Generate a random permutation of indices [0, min(90, n-i)-1]
			for j := 0; j < 15 && j < len(possibleDeps); j++ { // Each node has up to 15 random dependencies
				targetIndex := i + possibleDeps[j]
				shouldCycle := rand.Float64() < cycleProbability
				if targetIndex != i { // Avoid self-dependency and control cycle creation
					v := max(targetIndex-rand.Intn(100), 0)
					if shouldCycle && v != i {
						err := nodes[i].SetDependency(storage, nodes[v])
						assert.NoError(t, err)
					} else {
						err := nodes[i].SetDependency(storage, nodes[targetIndex])
						assert.NoError(t, err)
					}

				}
			}
		}

		// Precompute expected results for QueryDependentsNoCache and QueryDependenciesNoCache
		for _, node := range nodes {
			dependents, err := node.QueryDependentsNoCache(storage)
			assert.NoError(t, err)
			expectedDependents[node.ID] = dependents.ToArray()

			dependencies, err := node.QueryDependenciesNoCache(storage)
			assert.NoError(t, err)
			expectedDependencies[node.ID] = dependencies.ToArray()
		}

		start := time.Now()

		// Cache the current state
		err := Cache(storage)
		if err != nil {
			t.Fatal(err)
		}

		assert.NoError(t, err)

		t.Logf("Cache took %v for n = %v", time.Since(start), n)

		// Benchmark QueryDependents, QueryDependencies and Cache
		for _, node := range nodes {
			dependents, err := node.QueryDependents(storage)
			assert.NoError(t, err)
			depArr := []uint32{}
			if dependents != nil {
				depArr = dependents.ToArray()
			}
			assert.Equal(t, expectedDependents[node.ID], depArr, fmt.Sprintf("Dependents of node %v", node.ID))

			dependencies, err := node.QueryDependencies(storage)
			assert.NoError(t, err)
			depArr = []uint32{}
			if dependencies != nil {
				depArr = dependencies.ToArray()
			}
			assert.Equal(t, expectedDependencies[node.ID], depArr, fmt.Sprintf("Dependencies of node %v", node.ID))
		}
	}
}

func TestRandomGraphDependenciesNoCircles(t *testing.T) {
	tests := []int{1000}
	for _, n := range tests {
		storage := NewMockStorage()
		nodes := make([]*Node, n)
		expectedDependents := make(map[uint32][]uint32)
		expectedDependencies := make(map[uint32][]uint32)

		// Create nodes and set dependencies

		for i := 0; i < n; i++ {
			node, err := AddNode(storage, fmt.Sprintf("type %d", i+1), fmt.Sprintf("metadata %d", i), fmt.Sprintf("name %d", i+1))
			assert.NoError(t, err)
			nodes[i] = node
		}

		m := map[int][]int{}

		for i := 0; i < n; i++ {
			possibleDeps := rand.Perm(n - i - 1)               // Generate a random permutation of indices [0, n-i-2]
			for j := 0; j < 15 && j < len(possibleDeps); j++ { // Each node has up to 10 random dependencies
				targetIndex := possibleDeps[j] + i + 1
				if targetIndex < n {
					err := nodes[i].SetDependency(storage, nodes[targetIndex])
					assert.NoError(t, err)
					m[int(nodes[i].ID)] = append(m[int(nodes[i].ID)], int(nodes[targetIndex].ID))
				}
			}
		}

		// Precompute expected results for QueryDependentsNoCache and QueryDependenciesNoCache
		for _, node := range nodes {
			dependents, err := node.QueryDependentsNoCache(storage)
			assert.NoError(t, err)
			expectedDependents[node.ID] = dependents.ToArray()

			dependencies, err := node.QueryDependenciesNoCache(storage)
			assert.NoError(t, err)
			expectedDependencies[node.ID] = dependencies.ToArray()
		}

		start := time.Now()

		// Cache the current state
		err := Cache(storage)
		if err != nil {
			t.Fatal(err)
		}

		assert.NoError(t, err)

		t.Logf("Cache took %v for n = %v", time.Since(start), n)

		// Benchmark QueryDependents, QueryDependencies and Cache
		for _, node := range nodes {
			dependents, err := node.QueryDependents(storage)
			assert.NoError(t, err)
			depArr := []uint32{}
			if dependents != nil {
				depArr = dependents.ToArray()
			}
			assert.Equal(t, expectedDependents[node.ID], depArr, fmt.Sprintf("Dependents of node %v", node.ID))

			dependencies, err := node.QueryDependencies(storage)
			assert.NoError(t, err)
			depArr = []uint32{}
			if dependencies != nil {
				depArr = dependencies.ToArray()
			}
			assert.Equal(t, expectedDependencies[node.ID], depArr, fmt.Sprintf("Dependencies of node %v", node.ID))
		}
	}
}

func TestComplexCircularDependency(t *testing.T) {
	storage := NewMockStorage()
	nodes := make([]*Node, 13)
	var err error

	// Create nodes
	for i := 0; i < 13; i++ {
		nodes[i], err = AddNode(storage, fmt.Sprintf("type %d", i+1), fmt.Sprintf("metadata %d", i), fmt.Sprintf("name %d", i+1))
		assert.NoError(t, err, "Expected no error")
	}

	// Create circular dependencies like figure 8s
	// Circle 1: node0 -> node1 -> node2 -> node0
	err = nodes[0].SetDependency(storage, nodes[1])
	assert.NoError(t, err)
	err = nodes[1].SetDependency(storage, nodes[2])
	assert.NoError(t, err)
	err = nodes[2].SetDependency(storage, nodes[0])
	assert.NoError(t, err)

	// Circle 2: node3 -> node4 -> node5 -> node3
	err = nodes[3].SetDependency(storage, nodes[4])
	assert.NoError(t, err)
	err = nodes[4].SetDependency(storage, nodes[5])
	assert.NoError(t, err)
	err = nodes[5].SetDependency(storage, nodes[3])
	assert.NoError(t, err)

	// Figure 8 linking Circle 1 and Circle 2: node2 -> node3
	err = nodes[2].SetDependency(storage, nodes[3])
	assert.NoError(t, err)

	// Additional circle: node6 -> node7 -> node8 -> node9 -> node6
	err = nodes[6].SetDependency(storage, nodes[7])
	assert.NoError(t, err)
	err = nodes[7].SetDependency(storage, nodes[8])
	assert.NoError(t, err)
	err = nodes[8].SetDependency(storage, nodes[9])
	assert.NoError(t, err)
	err = nodes[9].SetDependency(storage, nodes[6])
	assert.NoError(t, err)

	// Linking node9 to node1 to form another figure 8 between Circle 1 and the additional circle
	err = nodes[9].SetDependency(storage, nodes[1])
	assert.NoError(t, err)

	// Additional independent circle: node10 -> node11 -> node12 -> node10
	err = nodes[10].SetDependency(storage, nodes[11])
	assert.NoError(t, err)
	err = nodes[11].SetDependency(storage, nodes[12])
	assert.NoError(t, err)
	err = nodes[12].SetDependency(storage, nodes[10])
	assert.NoError(t, err)

	if err := Cache(storage); err != nil {
		t.Fatal(err)
	}

	// Test QueryDependents and QueryDependencies for complex circular dependencies
	for _, node := range nodes {
		dependents, err := node.QueryDependents(storage)
		assert.NoError(t, err, "Expected no error when querying cached dependents")
		dependentsNoCache, err := node.QueryDependentsNoCache(storage)
		assert.NoError(t, err, "Expected no error when querying non-cached dependents")
		assert.Equal(t, dependentsNoCache.ToArray(), dependents.ToArray(), "Cached and non-cached dependents should match")

		dependencies, err := node.QueryDependencies(storage)
		assert.NoError(t, err, "Expected no error when querying cached dependencies")
		dependenciesNoCache, err := node.QueryDependenciesNoCache(storage)
		assert.NoError(t, err, "Expected no error when querying non-cached dependencies")
		assert.Equal(t, dependenciesNoCache.ToArray(), dependencies.ToArray(), "Cached and non-cached dependencies should match")
	}
	assert.NoError(t, err)
}

func TestIntermediateSimpleCircles(t *testing.T) {
	storage := NewMockStorage()
	nodes := make([]*Node, 6)
	var err error

	// Create nodes
	for i := 0; i < 6; i++ {
		nodes[i], err = AddNode(storage, fmt.Sprintf("type %d", i+1), fmt.Sprintf("metadata %d", i), fmt.Sprintf("name %d", i+1))
		assert.NoError(t, err, "Expected no error")
	}

	// Circle 1: node0 -> node1 -> node2 -> node0
	err = nodes[0].SetDependency(storage, nodes[1])
	assert.NoError(t, err)
	err = nodes[1].SetDependency(storage, nodes[2])
	assert.NoError(t, err)
	err = nodes[2].SetDependency(storage, nodes[0])
	assert.NoError(t, err)

	// Circle 2: node3 -> node4 -> node5 -> node3
	err = nodes[3].SetDependency(storage, nodes[4])
	assert.NoError(t, err)
	err = nodes[4].SetDependency(storage, nodes[5])
	assert.NoError(t, err)
	err = nodes[5].SetDependency(storage, nodes[3])
	assert.NoError(t, err)

	// Linking Circle 1 and Circle 2
	err = nodes[2].SetDependency(storage, nodes[3])
	assert.NoError(t, err)
	// err = nodes[5].SetDependency(storages, nodes[0])
	// assert.NoError(t, err)

	if err := Cache(storage); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)

	// Test QueryDependents and QueryDependencies for intermediate simple circles
	for _, node := range nodes {
		dependents, err := node.QueryDependents(storage)
		assert.NoError(t, err, "Expected no error when querying cached dependents")
		dependentsNoCache, err := node.QueryDependentsNoCache(storage)
		assert.NoError(t, err, "Expected no error when querying non-cached dependents")
		assert.Equal(t, dependentsNoCache.ToArray(), dependents.ToArray(), "Cached and non-cached dependents should match")

		dependencies, err := node.QueryDependencies(storage)
		assert.NoError(t, err, "Expected no error when querying cached dependencies")
		dependenciesNoCache, err := node.QueryDependenciesNoCache(storage)
		assert.NoError(t, err, "Expected no error when querying non-cached dependencies")
		assert.Equal(t, dependenciesNoCache.ToArray(), dependencies.ToArray(), "Cached and non-cached dependencies should match")
	}
}

func TestSimpleCircle(t *testing.T) {
	storage := NewMockStorage()
	nodes := make([]*Node, 3)
	var err error

	// Create nodes
	for i := 0; i < 3; i++ {
		nodes[i], err = AddNode(storage, fmt.Sprintf("type %d", i+1), fmt.Sprintf("metadata %d", i), fmt.Sprintf("name %d", i+1))
		assert.NoError(t, err, "Expected no error")
	}

	// Simple Circle: node0 -> node1 -> node2 -> node0
	err = nodes[0].SetDependency(storage, nodes[1])
	assert.NoError(t, err)
	err = nodes[1].SetDependency(storage, nodes[2])
	assert.NoError(t, err)
	err = nodes[2].SetDependency(storage, nodes[0])
	assert.NoError(t, err)

	if err := Cache(storage); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)

	// Test QueryDependents and QueryDependencies for simple circle
	for _, node := range nodes {
		dependents, err := node.QueryDependents(storage)
		assert.NoError(t, err, "Expected no error when querying cached dependents")
		dependentsNoCache, err := node.QueryDependentsNoCache(storage)
		assert.NoError(t, err, "Expected no error when querying non-cached dependents")
		assert.Equal(t, dependentsNoCache.ToArray(), dependents.ToArray(), "Cached and non-cached dependents should match")

		dependencies, err := node.QueryDependencies(storage)
		assert.NoError(t, err, "Expected no error when querying cached dependencies")
		dependenciesNoCache, err := node.QueryDependenciesNoCache(storage)
		assert.NoError(t, err, "Expected no error when querying non-cached dependencies")
		assert.Equal(t, dependenciesNoCache.ToArray(), dependencies.ToArray(), "Cached and non-cached dependencies should match")
	}
}

// TestCacheErrors tests the Cache function for various error conditions.
func TestCacheErrors(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockStorage)
		wantErr   string
	}{
		{
			name:      "ToBeCached error",
			setupMock: func(m *MockStorage) { m.ToBeCachedErr = fmt.Errorf("ToBeCached error") },
			wantErr:   "error getting uncached nodes: ToBeCached error",
		},
		{
			name:      "GetAllKeys error",
			setupMock: func(m *MockStorage) { m.GetAllKeysErr = fmt.Errorf("GetAllKeys error") },
			wantErr:   "error getting keys: GetAllKeys error",
		},
		{
			name:      "GetNodes error",
			setupMock: func(m *MockStorage) { m.GetNodesErr = fmt.Errorf("GetNodes error") },
			wantErr:   "error getting all nodes: GetNodes error",
		},
		{
			name:      "SaveCaches error",
			setupMock: func(m *MockStorage) { m.SaveCachesErr = fmt.Errorf("SaveCaches error") },
			wantErr:   "error saving caches: SaveCaches error",
		},
		{
			name:      "ClearCacheStack error",
			setupMock: func(m *MockStorage) { m.ClearCacheStackErr = fmt.Errorf("ClearCacheStack error") },
			wantErr:   "ClearCacheStack error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage()
			nodes := make([]*Node, 3)

			// Create nodes
			for i := 0; i < 3; i++ {
				var err error
				nodes[i], err = AddNode(mockStorage, fmt.Sprintf("type %d", i+1), fmt.Sprintf("metadata %d", i), fmt.Sprintf("name %d", i+1))
				assert.NoError(t, err)
			}

			// Create circle: node0 -> node1 -> node2 -> node0
			for i := 0; i < 3; i++ {
				err := nodes[i].SetDependency(mockStorage, nodes[(i+1)%3])
				assert.NoError(t, err)
			}

			tt.setupMock(mockStorage)
			err := Cache(mockStorage)
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("Expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}
