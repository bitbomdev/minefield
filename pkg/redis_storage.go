package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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
	data, err := json.Marshal(node)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), fmt.Sprintf("node:%d", node.id), data, 0).Err()
}

func (r *RedisStorage[T]) GetNode(id uint32) (*Node[T], error) {
	data, err := r.client.Get(context.Background(), fmt.Sprintf("node:%d", id)).Result()
	if err != nil {
		return nil, err
	}
	var node Node[T]
	if err := json.Unmarshal([]byte(data), &node); err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *RedisStorage[T]) GetAllKeys() ([]uint32, error) {
	keys, err := r.client.Keys(context.Background(), "node:*").Result()
	if err != nil {
		return nil, err
	}

	returnedKeys := make([]uint32, len(keys))

	for i, k := range keys {
		n, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		returnedKeys[i] = uint32(n)
	}

	return returnedKeys, nil
}

func (r *RedisStorage[T]) SetDependency(nodeID, neighborID uint32) error {
	node, err := r.GetNode(nodeID)
	if err != nil {
		return err
	}
	neighbor, err := r.GetNode(neighborID)
	if err != nil {
		return err
	}
	return node.SetDependency(r, neighbor)
}

func (r *RedisStorage[T]) QueryDependents(nodeID uint32) (*roaring.Bitmap, error) {
	node, err := r.GetNode(nodeID)
	if err != nil {
		return nil, err
	}
	return node.QueryDependents(r)
}

func (r *RedisStorage[T]) QueryDependencies(nodeID uint32) (*roaring.Bitmap, error) {
	node, err := r.GetNode(nodeID)
	if err != nil {
		return nil, err
	}
	return node.QueryDependencies(r)
}

func (r *RedisStorage[T]) GenerateID() (uint32, error) {
	id, err := r.client.Incr(context.Background(), "node_id_counter").Result()
	if err != nil {
		return 0, err
	}
	return uint32(id), nil
}
