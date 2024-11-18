package storages

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	NodeKeyPrefix  = "node:"
	NameToIDKey    = "name_to_id:"
	CacheKeyPrefix = "cache:"
	IDCounterKey   = "id_counter"
	CacheStackKey  = "to_be_cached"
)

// SetupRedisTestDB initializes a new SQLStorage with the given DSN.
func SetupSQLTestDB(dsn string) (*SQLStorage, error) {
	storage, err := NewSQLStorage(dsn, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLStorage: %w", err)
	}
	return storage, nil
}

// SetupRedisTestDB initializes a new RedisStorage with the given address.
func SetupRedisTestDB(addr string) *RedisStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	rdb.FlushDB(context.Background()) // Clear the database before each test
	return &RedisStorage{Client: rdb}
}
