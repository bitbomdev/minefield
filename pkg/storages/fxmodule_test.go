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
	// Base address for Redis
	var addr string
	var client *redis.Client
	var err error

	// Use a static port for Redis
	addr = "localhost:6379"
	client = redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		t.Skipf("Skipping testdata: Redis is not available at %s", addr)
	}

	// Create an fx.App for testing
	app := fx.New(
		NewRedisStorageModule(addr),
		fx.Invoke(func(storage graph.Storage) {
			assert.Implements(t, (*graph.Storage)(nil), storage, "RedisStorage should implement graph.Storage")

			// Perform additional assertions on the storage instance
			redisStorage, ok := storage.(*RedisStorage)
			assert.True(t, ok, "Expected storage to be of type *RedisStorage")
			assert.NotNil(t, redisStorage.Client, "RedisStorage Client should not be nil")
		}),
	)

	// Start and stop the fx.App
	assert.NoError(t, app.Start(context.Background()))
	assert.NoError(t, app.Stop(context.Background()))
}
