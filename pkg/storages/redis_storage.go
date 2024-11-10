package storages

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bitbomdev/minefield/pkg/graph"
	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	Client *redis.Client
}

func NewRedisStorage(addr string) graph.Storage {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStorage{Client: rdb}
}

func (r *RedisStorage) GenerateID() (uint32, error) {
	id, err := r.Client.Incr(context.Background(), "id_counter").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	return uint32(id), nil
}

func (r *RedisStorage) SaveNode(node *graph.Node) error {
	data, err := node.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}
	if err := r.Client.Set(context.Background(), fmt.Sprintf("node:%d", node.ID), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save node data: %w", err)
	}
	if err := r.Client.Set(context.Background(), fmt.Sprintf("name_to_id:%s", node.Name), strconv.Itoa(int(node.ID)), 0).Err(); err != nil {
		return fmt.Errorf("failed to save node name to ID mapping: %w", err)
	}
	if err := r.AddNodeToCachedStack(node.ID); err != nil {
		return fmt.Errorf("failed to add node ID to to_be_cached set: %w", err)
	}
	return nil
}

func (r *RedisStorage) NameToID(name string) (uint32, error) {
	id, err := r.Client.Get(context.Background(), fmt.Sprintf("name_to_id:%s", name)).Result()
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
	data, err := r.Client.Get(ctx, fmt.Sprintf("node:%d", id)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node data for ID %d: %w", id, err)
	}
	var node graph.Node
	if err := node.UnmarshalJSON([]byte(data)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
	}
	return &node, nil
}

func (r *RedisStorage) GetNodesByGlob(pattern string) ([]*graph.Node, error) {
	// Use pattern matching for Redis keys
	keys, err := r.Client.Keys(context.Background(), fmt.Sprintf("name_to_id:%s", pattern)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by pattern %s: %w", pattern, err)
	}

	nodes := make([]*graph.Node, 0, len(keys))
	for _, key := range keys {
		// Extract the name from the key
		name := strings.TrimPrefix(key, "name_to_id:")

		// Get the ID using the name
		id, err := r.NameToID(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get ID for name %s: %w", name, err)
		}

		node, err := r.GetNode(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get node for ID %d: %w", id, err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (r *RedisStorage) GetAllKeys() ([]uint32, error) {
	keys, err := r.Client.Keys(context.Background(), "node:*").Result()
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
	return r.Client.Set(ctx, fmt.Sprintf("cache:%d", cache.ID), data, 0).Err()
}

func (r *RedisStorage) ToBeCached() ([]uint32, error) {
	ctx := context.Background()
	data, err := r.Client.LRange(ctx, "to_be_cached", 0, -1).Result()
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
	err := r.Client.RPush(ctx, "to_be_cached", nodeID).Err()
	if err != nil {
		return fmt.Errorf("failed to add node %d to cached stack: %w", nodeID, err)
	}
	return nil
}

func (r *RedisStorage) ClearCacheStack() error {
	ctx := context.Background()
	err := r.Client.Del(ctx, "to_be_cached").Err()
	if err != nil {
		return fmt.Errorf("failed to clear cache stack: %w", err)
	}
	return nil
}

func (r *RedisStorage) GetCache(nodeID uint32) (*graph.NodeCache, error) {
	ctx := context.Background()
	data, err := r.Client.Get(ctx, fmt.Sprintf("cache:%d", nodeID)).Result()
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
	pipe := r.Client.Pipeline()

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
	pipe := r.Client.Pipeline()

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

func (r *RedisStorage) GetCaches(ids []uint32) (map[uint32]*graph.NodeCache, error) {
	ctx := context.Background()
	pipe := r.Client.Pipeline()

	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.Get(ctx, fmt.Sprintf("cache:%d", id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get caches: %w", err)
	}

	caches := make(map[uint32]*graph.NodeCache, len(ids))
	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip missing caches
		} else if err != nil {
			return nil, fmt.Errorf("failed to get cache data for ID %d: %w", ids[i], err)
		}

		var cache graph.NodeCache
		if err := cache.UnmarshalJSON([]byte(data)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
		}
		caches[ids[i]] = &cache
	}

	return caches, nil
}

func (r *RedisStorage) RemoveAllCaches() error {
	ctx := context.Background()
	var cursor uint64
	var err error

	for {
		var keys []string
		keys, cursor, err = r.Client.Scan(ctx, cursor, "cache:*", 1000).Result()
		if err != nil {
			return fmt.Errorf("failed to scan cache keys: %w", err)
		}

		if len(keys) > 0 {
			pipe := r.Client.Pipeline()

			// Extract IDs and add them to the cache stack
			for _, key := range keys {
				id := strings.TrimPrefix(key, "cache:")
				pipe.RPush(ctx, "to_be_cached", id)
			}

			// Delete the cache entries
			pipe.Unlink(ctx, keys...)

			_, err = pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to process cache keys: %w", err)
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

func (r *RedisStorage) AddOrUpdateCustomData(tag, key string, datakey string, data []byte) error {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", tag, key)

	// Use HSet to add or update the field in the hash
	err := r.Client.HSet(ctx, redisKey, datakey, data).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}

	return nil
}

// GetCustomData gets data from the database.
func (r *RedisStorage) GetCustomData(tag, key string) (map[string][]byte, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", tag, key)

	data, err := r.Client.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get data from DB: %w", err)
	}

	result := make(map[string][]byte, len(data))
	for field, value := range data {
		result[field] = []byte(value)
	}

	return result, nil
}
