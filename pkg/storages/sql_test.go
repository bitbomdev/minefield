package storages

import (
	"fmt"
	"os"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/stretchr/testify/assert"
)

// setupTestDB initializes a new SQLStorage with the given DSN.
func setupTestDB(dsn string) (*SQLStorage, error) {
	storage, err := NewSQLStorage(dsn, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLStorage: %w", err)
	}
	return storage, nil
}

// TestGenerateID_InMemory tests the GenerateID method using an in-memory SQLite database.
func TestSQLGenerateID_InMemory(t *testing.T) {
	storage, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	const numIDs = 10
	ids := make(map[uint32]bool)

	for i := 1; i <= numIDs; i++ {
		id, err := storage.GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed at iteration %d: %v", i, err)
		}
		if id != uint32(i) {
			t.Errorf("Expected ID %d, got %d", i, id)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}
}

// TestGenerateID_FileBased tests the GenerateID method using a file-based SQLite database.
func TestGenerateID_FileBased(t *testing.T) {
	// Create a temporary file for the SQLite database
	tempDB := "test_generate_id.db"
	defer os.Remove(tempDB) // Clean up after the test

	storage, err := setupTestDB(tempDB)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	const numIDs = 5
	ids := make(map[uint32]bool)

	for i := 1; i <= numIDs; i++ {
		id, err := storage.GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed at iteration %d: %v", i, err)
		}
		if id != uint32(i) {
			t.Errorf("Expected ID %d, got %d", i, id)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}

	// Re-initialize the storage to ensure persistence
	storage, err = setupTestDB(tempDB)
	if err != nil {
		t.Fatalf("Re-setup failed: %v", err)
	}

	// Generate additional IDs and ensure they continue from the last value
	for i := numIDs + 1; i <= numIDs*2; i++ {
		id, err := storage.GenerateID()
		if err != nil {
			t.Fatalf("GenerateID failed at iteration %d: %v", i, err)
		}
		if id != uint32(i) {
			t.Errorf("Expected ID %d, got %d", i, id)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %d", id)
		}
		ids[id] = true
	}
}

// TestGenerateID_InvalidStorage tests GenerateID method with an invalid storage setup.
func TestSQLSaveNode(t *testing.T) {
	s, err := setupTestDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	node := &graph.Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err = s.SaveNode(node)
	assert.NoError(t, err)

	// Verify node data is saved
	savedNode, err := s.GetNode(node.ID)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, savedNode.ID)
	assert.Equal(t, node.Name, savedNode.Name)
}
func TestSQLGetNodes(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	// Add test data
	node1 := &graph.Node{ID: 1, Name: "test_node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "test_node2", Children: roaring.New(), Parents: roaring.New()}
	node3 := &graph.Node{ID: 3, Name: "test_node3", Children: roaring.New(), Parents: roaring.New()}
	node2.Parents.Add(1)
	node1.Children.Add(2)
	node3.Parents.Add(1)
	node1.Children.Add(3)
	err = s.SaveNode(node1)
	assert.NoError(t, err)
	err = s.SaveNode(node2)
	assert.NoError(t, err)
	err = s.SaveNode(node3)
	assert.NoError(t, err)

	// Test GetNodes
	nodes, err := s.GetNodes([]uint32{1, 2})
	assert.NoError(t, err)
	assert.NotNil(t, nodes[1])
	assert.Equal(t, "test_node1", nodes[1].Name)
	assert.NotNil(t, nodes[2])
	assert.Equal(t, "test_node2", nodes[2].Name)
	assert.Equal(t, node1.ID, nodes[1].ID)
	assert.Equal(t, node2.ID, nodes[2].ID)
}

func TestSQLNameToID(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	node := &graph.Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err = s.SaveNode(node)
	assert.NoError(t, err)

	id, err := s.NameToID(node.Name)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, id)
}
func TestSQLGetAllKeys(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	node1 := &graph.Node{ID: 1, Name: "node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "node2", Children: roaring.New(), Parents: roaring.New()}
	err = s.SaveNode(node1)
	assert.NoError(t, err)
	err = s.SaveNode(node2)
	assert.NoError(t, err)

	keys, err := s.GetAllKeys()
	fmt.Println(keys)
	assert.NoError(t, err)
	assert.Contains(t, keys, node1.ID)
	assert.Contains(t, keys, node2.ID)
}

func TestSQLSaveCache(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	cache := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = s.SaveCache(cache)
	assert.NoError(t, err)

	savedCache, err := s.GetCache(cache.ID)
	assert.NoError(t, err)
	assert.Equal(t, cache.ID, savedCache.ID)
}
func TestSQLToBeCached(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	nodeID := uint32(1)
	err = s.AddNodeToCachedStack(nodeID)
	assert.NoError(t, err)

	toBeCached, err := s.ToBeCached()
	assert.NoError(t, err)
	assert.Contains(t, toBeCached, nodeID)
}
func TestSQLClearCacheStack(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	nodeID := uint32(1)
	err = s.AddNodeToCachedStack(nodeID)
	assert.NoError(t, err)

	err = s.ClearCacheStack()
	assert.NoError(t, err)

	toBeCached, err := s.ToBeCached()
	assert.NoError(t, err)
	assert.NotContains(t, toBeCached, nodeID)
}
func TestSQLSaveCaches(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	cache1 := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	cache2 := &graph.NodeCache{ID: 2, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = s.SaveCaches([]*graph.NodeCache{cache1, cache2})
	assert.NoError(t, err)

	// Verify caches saved
	savedCache1, err := s.GetCache(1)
	assert.NoError(t, err)
	assert.Equal(t, cache1.ID, savedCache1.ID)
	savedCache2, err := s.GetCache(2)
	assert.NoError(t, err)
	assert.Equal(t, cache2.ID, savedCache2.ID)
}
func TestSQLGetCaches(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	cache1 := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	cache2 := &graph.NodeCache{ID: 2, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = s.SaveCaches([]*graph.NodeCache{cache1, cache2})
	assert.NoError(t, err)

	// Test GetCaches
	caches, err := s.GetCaches([]uint32{1, 2})
	assert.NoError(t, err)
	assert.NotNil(t, caches[1])
	assert.Equal(t, cache1.ID, caches[1].ID)
	assert.NotNil(t, caches[2])
	assert.Equal(t, cache2.ID, caches[2].ID)
}
func TestSQLRemoveAllCaches(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	cache1 := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	cache2 := &graph.NodeCache{ID: 2, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = s.SaveCaches([]*graph.NodeCache{cache1, cache2})
	assert.NoError(t, err)

	// Test RemoveAllCaches
	err = s.RemoveAllCaches()
	assert.NoError(t, err)

	// Verify caches removed
	caches, err := s.GetCaches([]uint32{1, 2})
	assert.NoError(t, err)
	assert.Nil(t, caches[1])
	assert.Nil(t, caches[2])
}
func TestSQLAddAndGetDataToDB(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	err = s.AddOrUpdateCustomData("test_tag", "test_key1", "test_data1", []byte("test_data1"))
	assert.Error(t, err)
}

func TestSQLGetAllKeysByGlob(t *testing.T) {
	s, err := setupTestDB("file::memory:")
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	node1 := &graph.Node{ID: 1, Name: "node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "node2", Children: roaring.New(), Parents: roaring.New()}
	err = s.SaveNode(node1)
	assert.NoError(t, err)
	err = s.SaveNode(node2)
	assert.NoError(t, err)

	nodes, err := s.GetNodesByGlob("node%")
	assert.NoError(t, err)
	assert.Equal(t, node1.ID, nodes[0].ID)
}
