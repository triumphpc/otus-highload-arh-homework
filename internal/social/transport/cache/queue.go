package cachewarmer

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	stream string
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{
		client: client,
		stream: "cache_warm_tasks",
	}
}

func (q *RedisQueue) Push(ctx context.Context, task WarmTask) error {
	_, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: q.stream,
		Values: map[string]interface{}{
			"user_id": task.UserID,
		},
	}).Result()

	return err
}
