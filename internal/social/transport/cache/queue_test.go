package cachewarmer

import (
	"context"
	"errors"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisQueue_ProcessTasks(t *testing.T) {
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

	// Создаем очередь
	queue := NewRedisQueue(client)

	// Создаем consumer group
	_, err = client.XGroupCreateMkStream(ctx, "cache_warm_tasks", "cache_workers", "0").Result()
	require.NoError(t, err)

	// Отправляем тестовые задачи
	testUserIDs := []int{1, 2, 3, 4, 5}
	for _, id := range testUserIDs {
		err := queue.Push(ctx, WarmTask{UserID: id})
		require.NoError(t, err)
	}

	// Обрабатываем задачи вручную (имитация работы воркера)
	processedCount := 0
	for i := 0; i < len(testUserIDs); i++ {
		result, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    "cache_workers",
			Consumer: "test_consumer",
			Streams:  []string{"cache_warm_tasks", ">"},
			Count:    1,
			Block:    2 * time.Second,
		}).Result()

		if errors.Is(err, redis.Nil) {
			continue // Нет новых сообщений
		}
		require.NoError(t, err)

		for _, msg := range result[0].Messages {
			userIDStr := msg.Values["user_id"].(string)
			authorID, err := strconv.Atoi(userIDStr)
			require.NoError(t, err)
			log.Print(authorID)

			// Подтверждаем обработку
			_, err = client.XAck(ctx, "cache_warm_tasks", "cache_workers", msg.ID).Result()
			assert.NoError(t, err)

			processedCount++
		}
	}

	// Проверяем что нет pending сообщений
	groupInfo, err := client.XInfoGroups(ctx, "cache_warm_tasks").Result()
	require.NoError(t, err)
	for _, group := range groupInfo {
		if group.Name == "cache_workers" {
			assert.Equal(t, int64(0), group.Pending, "Should have no pending messages")
			break
		}
	}
}
