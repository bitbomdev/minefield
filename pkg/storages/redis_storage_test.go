package storages

import (
	"context"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {

	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	id, err := r.GenerateID()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)
}

func TestSaveNode(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	node := &graph.Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err = r.SaveNode(node)
	assert.NoError(t, err)

	// Verify node data is saved
	savedNode, err := r.GetNode(node.ID)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, savedNode.ID)
	assert.Equal(t, node.Name, savedNode.Name)
}

func TestNameToID(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	node := &graph.Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err = r.SaveNode(node)
	assert.NoError(t, err)

	id, err := r.NameToID(node.Name)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, id)
}

func TestGetAllKeys(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	node1 := &graph.Node{ID: 1, Name: "node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "node2", Children: roaring.New(), Parents: roaring.New()}
	err = r.SaveNode(node1)
	assert.NoError(t, err)
	err = r.SaveNode(node2)
	assert.NoError(t, err)

	keys, err := r.GetAllKeys()
	assert.NoError(t, err)
	assert.Contains(t, keys, node1.ID)
	assert.Contains(t, keys, node2.ID)
}

func TestSaveCache(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	cache := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = r.SaveCache(cache)
	assert.NoError(t, err)

	savedCache, err := r.GetCache(cache.ID)
	assert.NoError(t, err)
	assert.Equal(t, cache.ID, savedCache.ID)
}

func TestToBeCached(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	nodeID := uint32(1)
	err = r.AddNodeToCachedStack(nodeID)
	assert.NoError(t, err)

	toBeCached, err := r.ToBeCached()
	assert.NoError(t, err)
	assert.Contains(t, toBeCached, nodeID)
}

func TestClearCacheStack(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	nodeID := uint32(1)
	err = r.AddNodeToCachedStack(nodeID)
	assert.NoError(t, err)

	err = r.ClearCacheStack()
	assert.NoError(t, err)

	toBeCached, err := r.ToBeCached()
	assert.NoError(t, err)
	assert.NotContains(t, toBeCached, nodeID)
}

func TestGetNodes(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	// Add test data
	node1 := &graph.Node{ID: 1, Name: "test_node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "test_node2", Children: roaring.New(), Parents: roaring.New()}
	err = r.SaveNode(node1)
	assert.NoError(t, err)
	err = r.SaveNode(node2)
	assert.NoError(t, err)

	// Test GetNodes
	nodes, err := r.GetNodes([]uint32{1, 2})
	assert.NoError(t, err)
	assert.NotNil(t, nodes[1])
	assert.Equal(t, "test_node1", nodes[1].Name)
	assert.NotNil(t, nodes[2])
	assert.Equal(t, "test_node2", nodes[2].Name)
}

func TestSaveCaches(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	cache1 := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	cache2 := &graph.NodeCache{ID: 2, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = r.SaveCaches([]*graph.NodeCache{cache1, cache2})
	assert.NoError(t, err)

	// Verify caches saved
	savedCache1, err := r.GetCache(1)
	assert.NoError(t, err)
	assert.Equal(t, cache1.ID, savedCache1.ID)
	savedCache2, err := r.GetCache(2)
	assert.NoError(t, err)
	assert.Equal(t, cache2.ID, savedCache2.ID)
}

func TestGetCaches(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	cache1 := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	cache2 := &graph.NodeCache{ID: 2, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = r.SaveCaches([]*graph.NodeCache{cache1, cache2})
	assert.NoError(t, err)

	// Test GetCaches
	caches, err := r.GetCaches([]uint32{1, 2})
	assert.NoError(t, err)
	assert.NotNil(t, caches[1])
	assert.Equal(t, cache1.ID, caches[1].ID)
	assert.NotNil(t, caches[2])
	assert.Equal(t, cache2.ID, caches[2].ID)
}

func TestRemoveAllCaches(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	cache1 := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	cache2 := &graph.NodeCache{ID: 2, AllParents: roaring.New(), AllChildren: roaring.New()}
	err = r.SaveCaches([]*graph.NodeCache{cache1, cache2})
	assert.NoError(t, err)

	// Test RemoveAllCaches
	err = r.RemoveAllCaches()
	assert.NoError(t, err)

	// Verify caches removed
	caches, err := r.GetCaches([]uint32{1, 2})
	assert.NoError(t, err)
	assert.Nil(t, caches[1])
	assert.Nil(t, caches[2])
}

func TestAddAndGetDataToDB(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	err = r.AddOrUpdateCustomData("test_tag", "test_key1", "test_data1", []byte("test_data1"))
	assert.NoError(t, err)
	err = r.AddOrUpdateCustomData("test_tag", "test_key1", "test_data2", []byte("test_data2"))
	assert.NoError(t, err)

	// Verify data added
	data, err := r.GetCustomData("test_tag", "test_key1")
	assert.NoError(t, err)

	t1, err := json.Marshal("test_data1")
	assert.NoError(t, err)
	assert.Contains(t, string(t1), string(data["test_data1"]))

	t2, err := json.Marshal("test_data2")
	assert.NoError(t, err)
	assert.Contains(t, string(t2), string(data["test_data2"]))
}

func TestGetNodesByGlob(t *testing.T) {
	r, err := SetupRedisTestDB(context.Background())
	assert.NoError(t, err)
	// Add test nodes
	node1 := &graph.Node{ID: 1, Name: "test_node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "test_node2", Children: roaring.New(), Parents: roaring.New()}
	node3 := &graph.Node{ID: 3, Name: "other_node", Children: roaring.New(), Parents: roaring.New()}
	err = r.SaveNode(node1)
	assert.NoError(t, err)
	err = r.SaveNode(node2)
	assert.NoError(t, err)
	err = r.SaveNode(node3)
	assert.NoError(t, err)

	// Test GetNodesByGlob with pattern "test_*"
	nodes, err := r.GetNodesByGlob("test_*")
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)

	// Verify the nodes returned match the expected nodes
	nodeNames := []string{nodes[0].Name, nodes[1].Name}
	assert.Contains(t, nodeNames, "test_node1")
	assert.Contains(t, nodeNames, "test_node2")

	// Test with a pattern that matches no nodes
	nodes, err = r.GetNodesByGlob("nonexistent_*")
	assert.NoError(t, err)
	assert.Len(t, nodes, 0)

	// Simulate an error by closing the Redis client
	r.Client.Close()
	_, err = r.GetNodesByGlob("test_*")
	assert.Error(t, err)
}
