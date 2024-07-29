package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/RoaringBitmap/roaring"
	"github.com/go-redis/redis/v8"
)

type RedisStorage[T any] struct {
	client *redis.Client
}

func NewRedisStorage[T any](addr string) *RedisStorage[T] {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStorage[T]{client: rdb}
}

func (r *RedisStorage[T]) SaveNode(node *Node[T]) error {
	data, err := node.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal node: %w", err)
	}
	if err := r.client.Set(context.Background(), fmt.Sprintf("node:%d", node.Id), data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save node data: %w", err)
	}
	if err := r.client.Set(context.Background(), fmt.Sprint("id_to_name:", node.Id), node.Name, 0).Err(); err != nil {
		return fmt.Errorf("failed to save node ID to name mapping: %w", err)
	}
	if err := r.client.Set(context.Background(), fmt.Sprint("name_to_id:", node.Name), strconv.Itoa(int(node.Id)), 0).Err(); err != nil {
		return fmt.Errorf("failed to save node name to ID mapping: %w", err)
	}
	return nil
}

func (r *RedisStorage[T]) GetNode(id uint32) (*Node[T], error) {
	data, err := r.client.Get(context.Background(), fmt.Sprintf("node:%d", id)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get node data: %w", err)
	}
	var node Node[T]
	if err := node.UnmarshalJSON([]byte(data)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node data: %w", err)
	}
	return &node, nil
}

func (r *RedisStorage[T]) GetAllKeys() ([]uint32, error) {
	keys, err := r.client.Keys(context.Background(), "node:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}

	returnedKeys := make([]uint32, len(keys))

	for i, k := range keys {
		arr := strings.Split(k, ":")
		n, err := strconv.Atoi(arr[1])
		if err != nil {
			return nil, fmt.Errorf("failed to convert key to integer: %w", err)
		}
		returnedKeys[i] = uint32(n)
	}

	return returnedKeys, nil
}

func (r *RedisStorage[T]) SaveCache(cache *NodeCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	if err := r.AddNodeToCachedStack(cache.nodeID); err != nil {
		return err
	}
	return r.client.Set(context.Background(), fmt.Sprintf("cacheHelper:%d", cache.nodeID), data, 0).Err()
}

// ToBeCached returns all nodes that haven't been cached
func (r *RedisStorage[T]) ToBeCached() ([]uint32, error) {
	data, err := r.client.Get(context.Background(), "toBeCached").Result()
	if err != nil {
		if strings.Contains(err.Error(), "redis: nil") {
			return []uint32{}, nil
		}
		return nil, err
	}
	var toBeCached []uint32
	if err := json.Unmarshal([]byte(data), &toBeCached); err != nil {
		return nil, err
	}
	return toBeCached, nil
}

func (r *RedisStorage[T]) AddNodeToCachedStack(id uint32) error {
	var toBeCached []uint32

	data, err := r.client.Get(context.Background(), "toBeCached").Result()
	if !errors.Is(err, redis.Nil) {
		return err
	} else if err == nil {
		if err := json.Unmarshal([]byte(data), &toBeCached); err != nil {
			return err
		}
	}

	toBeCached = append(toBeCached, id)

	setData, err := json.Marshal(toBeCached)
	if err != nil {
		return err
	}

	return r.client.Set(context.Background(), "toBeCached", setData, 0).Err()
}

func (r *RedisStorage[T]) ClearCacheStack() error {
	var toBeCached []uint32
	setData, err := json.Marshal(toBeCached)
	if err != nil {
		return err
	}

	return r.client.Set(context.Background(), "toBeCached", setData, 0).Err()
}

func (r *RedisStorage[T]) GetCache(id uint32) (*NodeCache, error) {
	data, err := r.client.Get(context.Background(), fmt.Sprintf("cacheHelper:%d", id)).Result()
	if err != nil {
		return nil, err
	}
	var nodeCache NodeCache
	if err := json.Unmarshal([]byte(data), &nodeCache); err != nil {
		return nil, err
	}
	return &nodeCache, nil
}

func (r *RedisStorage[T]) SetDependency(nodeID, neighborID uint32) error {
	node, err := r.GetNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}
	neighbor, err := r.GetNode(neighborID)
	if err != nil {
		return fmt.Errorf("failed to get neighbor node: %w", err)
	}
	return node.SetDependency(r, neighbor)
}

func (r *RedisStorage[T]) QueryDependents(nodeID uint32) (*roaring.Bitmap, error) {
	node, err := r.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	return node.QueryDependents(r)
}

func (r *RedisStorage[T]) QueryDependencies(nodeID uint32) (*roaring.Bitmap, error) {
	node, err := r.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	return node.QueryDependencies(r)
}

func (r *RedisStorage[T]) GenerateID() (uint32, error) {
	id, err := r.client.Incr(context.Background(), "node_id_counter").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	return uint32(id), nil
}

func (r *RedisStorage[T]) NameToID(name string) (uint32, error) {
	id, err := r.client.Get(context.Background(), fmt.Sprint("name_to_id:", name)).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ID from name: %w", err)
	}
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("failed to convert ID to integer: %w", err)
	}
	return uint32(n), nil
}

func (r *RedisStorage[T]) IDToName(id uint32) (string, error) {
	name, err := r.client.Get(context.Background(), fmt.Sprint("id_to_name:", id)).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get name from ID: %w", err)
	}
	return name, nil
}
