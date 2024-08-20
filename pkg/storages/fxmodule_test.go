package storages

import (
	"context"
	"testing"

	"github.com/bit-bom/minefield/pkg/graph"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestNewRedisStorageModule(t *testing.T) {
	setupTestRedis()
	// Create a test Redis address
	addr := "localhost:6379"

	// Create a Redis client to check if Redis is available
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	// Check if Redis is available
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		t.Skipf("Skipping test: Redis is not available at %s", addr)
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
