package pkg

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr string) *RedisStorage {
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

func (r *RedisStorage) SaveNode(node *Node) error {
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

func (r *RedisStorage) GetNode(id uint32) (*Node, error) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, fmt.Sprintf("node:%d", id)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node data for ID %d: %w", id, err)
	}
	var node Node
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

func (r *RedisStorage) SaveCache(cache *NodeCache) error {
	ctx := context.Background()
	data, err := cache.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	return r.client.Set(ctx, fmt.Sprintf("cache:%d", cache.nodeID), data, 0).Err()
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

func (r *RedisStorage) GetCache(nodeID uint32) (*NodeCache, error) {
	ctx := context.Background()
	data, err := r.client.Get(ctx, fmt.Sprintf("cache:%d", nodeID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache for node %d: %w", nodeID, err)
	}
	var cache NodeCache
	if err := cache.UnmarshalJSON([]byte(data)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}
	return &cache, nil
}

func (r *RedisStorage) GetNodes(ids []uint32) (map[uint32]*Node, error) {
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

	nodes := make(map[uint32]*Node, len(ids))
	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err == redis.Nil {
			continue // Skip missing nodes
		} else if err != nil {
			return nil, fmt.Errorf("failed to get node data for ID %d: %w", ids[i], err)
		}

		var node Node
		if err := node.UnmarshalJSON([]byte(data)); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
		}
		nodes[ids[i]] = &node
	}

	return nodes, nil
}
