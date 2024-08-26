package storages

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/go-redis/redis/v8"
)

const (
	idCounterKey   = "id_counter"
	nodeKeyPrefix  = "node:"
	cacheKeyPrefix = "cache:"
	nameToIDPrefix = "name_to_id:"
	toBeCachedKey  = "to_be_cached"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr string) graph.Storage {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStorage{client: rdb}
}

func (r *RedisStorage) GenerateID() (uint32, error) {
	id, err := r.client.Incr(context.Background(), "id_counter").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	return uint32(id), nil
}

// SaveNode saves a node to the Redis storage.
func (r *RedisStorage) SaveNode(node *graph.Node) error {
	data, err := node.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}
	if err := r.client.Set(context.Background(), fmt.Sprintf("node:%d", node.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save node data: %w", err)
	}
	if err := r.client.Set(context.Background(), fmt.Sprint("name_to_id:", node.Name), strconv.Itoa(int(node.ID)), 0).Err(); err != nil {
		return fmt.Errorf("failed to save node name to ID mapping: %w", err)
	}
	if err := r.AddNodeToCachedStack(node.ID); err != nil {
		return fmt.Errorf("failed to add node ID to to_be_cached set: %w", err)
	}
	return nil
}

func (r *RedisStorage) NameToID(name string) (uint32, error) {
	id, err := r.client.Get(context.Background(), fmt.Sprintf("name_to_id:%s", name)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ID for name %s: %w", name, err)
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("failed to convert ID to integer: %w", err)
	}
	return uint32(idInt), nil
}

func (r *RedisStorage) GetNode(id uint32) (*graph.Node, error) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, fmt.Sprintf("node:%d", id)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node data for ID %d: %w", id, err)
	}
	var node graph.Node
	if err := node.UnmarshalJSON([]byte(data)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
	}
	return &node, nil
}

func (r *RedisStorage) GetAllKeys() ([]uint32, error) {
	keys, err := r.client.Keys(context.Background(), "node:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}
	var result []uint32
	for _, key := range keys {
		id, err := strconv.ParseUint(strings.TrimPrefix(key, "node:"), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key %s: %w", key, err)
		}
		result = append(result, uint32(id))
	}
	return result, nil
}

func (r *RedisStorage) SaveCache(cache *graph.NodeCache) error {
	ctx := context.Background()
	data, err := cache.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	return r.client.Set(ctx, fmt.Sprintf("cache:%d", cache.ID), data, 0).Err()
}

func (r *RedisStorage) ToBeCached() ([]uint32, error) {
	ctx := context.Background()
	// Use SMEMBERS to get all members of the set
	data, err := r.client.SMembers(ctx, "to_be_cached").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get to_be_cached data: %w", err)
	}

	result := make([]uint32, 0, len(data))
	for _, item := range data {
		id, err := strconv.ParseUint(item, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse item %s in to_be_cached: %w", item, err)
		}
		result = append(result, uint32(id))
	}

	return result, nil
}

func (r *RedisStorage) AddNodeToCachedStack(nodeID uint32) error {
	ctx := context.Background()
	err := r.client.SAdd(ctx, "to_be_cached", nodeID).Err()
	if err != nil {
		return fmt.Errorf("failed to add node %d to cached stack: %w", nodeID, err)
	}
	return nil
}

func (r *RedisStorage) ClearCacheStack() error {
	ctx := context.Background()
	err := r.client.Del(ctx, "to_be_cached").Err()
	if err != nil {
		return fmt.Errorf("failed to clear cache stack: %w", err)
	}
	return nil
}

func (r *RedisStorage) GetCache(nodeID uint32) (*graph.NodeCache, error) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, fmt.Sprintf("cache:%d", nodeID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache for node %d: %w", nodeID, err)
	}
	var cache graph.NodeCache
	if err := cache.UnmarshalJSON([]byte(data)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	return &cache, nil
}

func (r *RedisStorage) GetNodes(ids []uint32) (map[uint32]*graph.Node, error) {
	ctx := context.Background()
	pipe := r.client.Pipeline()

	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.Get(ctx, fmt.Sprintf("node:%d", id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	nodes := make(map[uint32]*graph.Node, len(ids))
	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip missing nodes
		} else if err != nil {
			return nil, fmt.Errorf("failed to get node data for ID %d: %w", ids[i], err)
		}

		var node graph.Node
		if err := node.UnmarshalJSON([]byte(data)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
		}
		nodes[ids[i]] = &node
	}

	return nodes, nil
}

func (r *RedisStorage) SaveCaches(caches []*graph.NodeCache) error {
	ctx := context.Background()
	pipe := r.client.Pipeline()

	for _, cache := range caches {
		data, err := cache.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal cache: %w", err)
		}
		pipe.Set(ctx, fmt.Sprintf("cache:%d", cache.ID), data, 0)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save caches: %w", err)
	}
	return nil
}

// SaveNodes saves nodes to the Redis storage.
// This will save using MSET, which is atomic, but it will not be atomic with respect to the cache.
func (r *RedisStorage) SaveNodes(nodes []*graph.Node) error {
	ctx := context.Background()
	pipe := r.client.Pipeline()

	nodeData := make(map[string]interface{})
	nameToIDData := make(map[string]interface{})
	toBeCached := make([]interface{}, len(nodes))

	for i, node := range nodes {
		data, err := node.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal node: %w", err)
		}
		nodeData[fmt.Sprintf("node:%d", node.ID)] = data
		nameToIDData[fmt.Sprintf("name_to_id:%s", node.Name)] = strconv.Itoa(int(node.ID))
		toBeCached[i] = node.ID
	}

	pipe.MSet(ctx, nodeData)
	pipe.MSet(ctx, nameToIDData)
	pipe.SAdd(ctx, "to_be_cached", toBeCached...)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save nodes: %w", err)
	}
	return nil
}

// BatchSaveNodes saves nodes to the Redis storage in a single batch operation.
// This function ensures atomicity and efficiency when saving multiple nodes.
func (r *RedisStorage) BatchSaveNodes(ctx context.Context, nodes []*graph.Node) error {
	/*
		BatchSaveNodes is a high-level function designed to save nodes to Redis in a single batch operation.
		The primary goal of this function is to ensure atomicity and efficiency when saving multiple nodes.

		Reasoning:
		1. **Atomicity**: By using Redis pipelines, we ensure that all operations are executed as a single unit. This reduces the risk of partial updates and ensures data consistency.
		2. **Efficiency**: Batch operations reduce the number of round trips to the Redis server, leading to better performance, especially when dealing with a large number of nodes.
		3. **ID Generation**: Each node needs a unique identifier. By incrementing a central counter (`idCounterKey`), we ensure that each node gets a unique ID.
		4. **Data Integrity**: The function ensures that the nodes are correctly saved with their unique IDs.

		Steps:
		1. Create a Redis pipeline to batch the commands.
		2. Generate unique IDs for each node by incrementing the `idCounterKey`.
		3. Execute the pipeline to get the new IDs.
		4. Assign the generated IDs to the nodes.
		5. Marshal the nodes to JSON format.
		6. Prepare the data for MSet (multi-set) operations, including node data and name-to-ID mappings.
		7. Use MSet to set multiple key-value pairs in Redis for nodes and name-to-ID mappings.
		8. Add the node IDs to the `to_be_cached` set in Redis.
		9. Execute the pipeline to perform the batch save operation.
		10. Return an error if the operation fails, otherwise return nil.
	*/

	pipe := r.client.Pipeline()

	// Generate IDs
	idCmds := make([]*redis.IntCmd, len(nodes))
	for i := range nodes {
		idCmds[i] = pipe.Incr(ctx, idCounterKey)
	}

	// Execute the pipeline to get the new IDs
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate IDs: %w", err)
	}

	// Assign IDs and prepare data for MSet
	nodeData := make(map[string]interface{})
	nameToIDData := make(map[string]interface{})
	toBeCached := make([]interface{}, len(nodes))

	for i, node := range nodes {
		node.ID = uint32(idCmds[i].Val())

		nodeJSON, err := node.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal node: %w %v", err, node.Name)
		}

		nodeData[fmt.Sprintf("%s%d", nodeKeyPrefix, node.ID)] = nodeJSON
		nameToIDData[fmt.Sprintf("%s%s", nameToIDPrefix, node.Name)] = strconv.Itoa(int(node.ID))
		toBeCached[i] = node.ID
	}

	// Use MSet to set multiple key-value pairs
	pipe.MSet(ctx, nodeData)
	pipe.MSet(ctx, nameToIDData)
	pipe.SAdd(ctx, toBeCachedKey, toBeCached...)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to save nodes: %w", err)
	}

	return nil
}
