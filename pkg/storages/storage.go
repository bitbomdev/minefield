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

// SetupSQLTestDB initializes a new SQLStorage with the given DSN.
func SetupSQLTestDB(dsn string) (*SQLStorage, error) {
	storage, err := NewSQLStorage(dsn, false)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLStorage: %w", err)
	}
	return storage, nil
}

// SetupRedisTestDB initializes a new RedisStorage with the given address.
func SetupRedisTestDB(ctx context.Context, addr string) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Verify connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Clear the database before each test
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to flush Redis database: %w", err)
	}

	return &RedisStorage{Client: rdb}, nil
}
