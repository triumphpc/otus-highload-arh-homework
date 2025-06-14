package cachewarmer

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testRedisContainer struct {
	testcontainers.Container
	URI string
}

func setupRedis(ctx context.Context) (*testRedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	uri := "redis://" + host + ":" + port.Port()

	return &testRedisContainer{Container: container, URI: uri}, nil
}

func TestCacheWarmer_SetAndGet(t *testing.T) {
	ctx := context.Background()

	// Запускаем тестовый Redis
	redisContainer, err := setupRedis(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	// Настраиваем клиент Redis
	opt, err := redis.ParseURL(redisContainer.URI)
	require.NoError(t, err)

	client := redis.NewClient(opt)
	defer client.Close()

	// Создаем CacheWarmer с mock для MessageQueue
	mockQueue := &mockMessageQueue{}
	warmer := New(mockQueue, client)

	t.Run("successful set and get", func(t *testing.T) {
		key := "test_key"
		value := map[string]interface{}{
			"field1": "value1",
			"field2": float64(42),
			"field3": true,
		}

		// Записываем значение в кэш
		err := warmer.Set(ctx, key, value, time.Minute)
		assert.NoError(t, err)

		// Читаем значение из кэша
		var result map[string]interface{}
		err = warmer.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("get non-existent key", func(t *testing.T) {
		var result map[string]interface{}
		err := warmer.Get(ctx, "non_existent_key", &result)
		assert.ErrorIs(t, err, ErrCacheMiss)
	})

	t.Run("invalid value type", func(t *testing.T) {
		key := "invalid_key"
		invalidValue := make(chan int) // Каналы нельзя маршалить в JSON

		err := warmer.Set(ctx, key, invalidValue, time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal value")

		// Попытка прочитать в несоответствующий тип
		err = warmer.Get(ctx, key, make(chan int))
		assert.Error(t, err)
	})
}

// mockMessageQueue реализует MessageQueue для тестов
type mockMessageQueue struct{}

func (m *mockMessageQueue) Push(ctx context.Context, task WarmTask) error {
	return nil // Просто заглушка, не используется в этих тестах
}
