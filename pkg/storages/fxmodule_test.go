package storages

import (
	"context"
	"fmt"
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestNewRedisStorageModule(t *testing.T) {
	setupTestRedis()
	// Base address for Redis
	var addr string
	var client *redis.Client
	var err error

	// Try different ports until a working one is found
	for port := 6379; port <= 6389; port++ {
		addr = fmt.Sprintf("localhost:%d", port)
		client = redis.NewClient(&redis.Options{Addr: addr})
		defer client.Close()

		_, err = client.Ping(context.Background()).Result()
		if err == nil {
			break
		} else {
			t.Logf("Redis not available at %s, trying next port...", addr)
		}
	}

	if err != nil {
		t.Skipf("Skipping test: Redis is not available on any tested port")
	}

	// Create an fx.App for testing
	app := fx.New(
		NewRedisStorageModule(addr),
		fx.Invoke(func(storage graph.Storage) {
			assert.Implements(t, (*graph.Storage)(nil), storage, "RedisStorage should implement graph.Storage")

			// Perform additional assertions on the storage instance
			redisStorage, ok := storage.(*RedisStorage)
			assert.True(t, ok, "Expected storage to be of type *RedisStorage")
			assert.NotNil(t, redisStorage.client, "RedisStorage client should not be nil")
		}),
	)

	// Start and stop the fx.App
	assert.NoError(t, app.Start(context.Background()))
	assert.NoError(t, app.Stop(context.Background()))
}
