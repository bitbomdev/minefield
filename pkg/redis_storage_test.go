package pkg

import (
	"context"
	"testing"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-redis/redis/v8"
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
	node := &Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
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
	node := &Node{ID: 1, Name: "test_node", Children: roaring.New(), Parents: roaring.New()}
	err := r.SaveNode(node)
	assert.NoError(t, err)

	id, err := r.NameToID(node.Name)
	assert.NoError(t, err)
	assert.Equal(t, node.ID, id)
}

func TestGetAllKeys(t *testing.T) {
	r := setupTestRedis()
	node1 := &Node{ID: 1, Name: "node1", Children: roaring.New(), Parents: roaring.New()}
	node2 := &Node{ID: 1, Name: "node2", Children: roaring.New(), Parents: roaring.New()}
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
	cache := &NodeCache{nodeID: 1, allParents: roaring.New(), allChildren: roaring.New()}
	err := r.SaveCache(cache)
	assert.NoError(t, err)

	savedCache, err := r.GetCache(cache.nodeID)
	assert.NoError(t, err)
	assert.Equal(t, cache.nodeID, savedCache.nodeID)
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
