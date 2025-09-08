package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const SystemTaskNameCounterRecalc = "system_task_counter_recalc"

type RedisQueue struct {
	client *redis.Client
	stream string
}

func NewRedisQueue(client *redis.Client) *RedisQueue {
	return &RedisQueue{
		client: client,
	}
}

func (q *RedisQueue) PushCounterRecalc(ctx context.Context, userID int64) error {
	_, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: SystemTaskNameCounterRecalc,
		Values: map[string]any{
			"user_id": userID,
		},
	}).Result()

	return err
}
