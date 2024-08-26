package storages

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/go-redis/redis/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func setupTestRedis() *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	rdb.FlushDB(context.Background()) // Clear the database before each test
	return &RedisStorage{client: rdb}
}

func TestGenerateID(t *testing.T) {
	r := setupTestRedis()
	id, err := r.GenerateID()
	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)
}

func TestSaveNode(t *testing.T) {
	r := setupTestRedis()
	node := &graph.Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err := r.SaveNode(node)
	assert.NoError(t, err)

	// Verify node data is saved
	savedNode, err := r.GetNode(node.ID)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, savedNode.ID)
	assert.Equal(t, node.Name, savedNode.Name)
}

func TestNameToID(t *testing.T) {
	r := setupTestRedis()
	node := &graph.Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err := r.SaveNode(node)
	assert.NoError(t, err)

	id, err := r.NameToID(node.Name)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, id)
}

func TestGetAllKeys(t *testing.T) {
	r := setupTestRedis()
	node1 := &graph.Node{ID: 1, Name: "node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &graph.Node{ID: 2, Name: "node2", Children: roaring.New(), Parents: roaring.New()}
	err := r.SaveNode(node1)
	assert.NoError(t, err)
	err = r.SaveNode(node2)
	assert.NoError(t, err)

	keys, err := r.GetAllKeys()
	assert.NoError(t, err)
	assert.Contains(t, keys, node1.ID)
	assert.Contains(t, keys, node2.ID)
}

func TestSaveCache(t *testing.T) {
	r := setupTestRedis()
	cache := &graph.NodeCache{ID: 1, AllParents: roaring.New(), AllChildren: roaring.New()}
	err := r.SaveCache(cache)
	assert.NoError(t, err)

	savedCache, err := r.GetCache(cache.ID)
	assert.NoError(t, err)
	assert.Equal(t, cache.ID, savedCache.ID)
}

func TestToBeCached(t *testing.T) {
	r := setupTestRedis()
	nodeID := uint32(1)
	err := r.AddNodeToCachedStack(nodeID)
	assert.NoError(t, err)

	toBeCached, err := r.ToBeCached()
	assert.NoError(t, err)
	assert.Contains(t, toBeCached, nodeID)
}

func TestClearCacheStack(t *testing.T) {
	r := setupTestRedis()
	nodeID := uint32(1)
	err := r.AddNodeToCachedStack(nodeID)
	assert.NoError(t, err)

	err = r.ClearCacheStack()
	assert.NoError(t, err)

	toBeCached, err := r.ToBeCached()
	assert.NoError(t, err)
	assert.NotContains(t, toBeCached, nodeID)
}

func TestGenerateSaveNodesAndCaches(t *testing.T) {
	r := setupTestRedis()
	defer r.client.Close()
	testSbom, err := os.ReadFile("../../test/kubernetes-sigs_kind.sbom.json")
	assert.NoError(t, err)

	// Prepare test data
	nodes := make([]*graph.Node, 10)
	for i := 0; i < 10; i++ {
		nodes[i] = &graph.Node{
			Name:     "Node" + strconv.Itoa(i+1),
			Metadata: testSbom,
			Type:     "Type" + strconv.Itoa(i+1),
			Children: roaring.New(),
			Parents:  roaring.New(),
		}
	}
	start := time.Now()
	// Call the method
	err = r.BatchSaveNodes(context.Background(), nodes)
	elapsed := time.Since(start)
	t.Logf("Time taken: %s", elapsed)
	assert.NoError(t, err)
	// Verify the results
	for _, node := range nodes {
		data, err := r.client.Get(context.Background(), "node:"+strconv.Itoa(int(node.ID))).Result()
		assert.NotZero(t, node.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		// Verify the structure and fields of the node
		var storedNode graph.Node
		err = json.Unmarshal([]byte(data), &storedNode)
		assert.NoError(t, err)
		if diff := cmp.Diff(node, &storedNode, cmpopts.IgnoreUnexported(roaring.Bitmap{}), cmpopts.IgnoreFields(graph.Node{}, "Metadata")); diff != "" {
			t.Errorf("Node mismatch (-want +got):\n%s", diff)
		}
	}

	for _, node := range nodes {
		id, err := r.client.Get(context.Background(), "name_to_id:"+node.Name).Result()
		assert.NoError(t, err)
		assert.Equal(t, strconv.Itoa(int(node.ID)), id)
	}

	toBeCached, err := r.client.SMembers(context.Background(), "to_be_cached").Result()
	assert.NoError(t, err)
	assert.Len(t, toBeCached, len(nodes))

	// Check for unexpected keys
	keys, err := r.client.Keys(context.Background(), "*").Result()
	assert.NoError(t, err)
	expectedKeys := make(map[string]bool)
	for _, node := range nodes {
		expectedKeys["node:"+strconv.Itoa(int(node.ID))] = true
		expectedKeys["name_to_id:"+node.Name] = true
	}
	expectedKeys["to_be_cached"] = true
	expectedKeys["id_counter"] = true // Add id_counter as a valid key

	for _, key := range keys {
		assert.True(t, expectedKeys[key], "Unexpected key found: %s", key)
	}
}
