package queue

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"otus-highload-arh-homework/internal/social/transport/service"
	"otus-highload-arh-homework/pkg/queue"

	"github.com/redis/go-redis/v9"
)

const groupName = "system_tasks_workers"

func StartCountersWorkers(ctx context.Context, client *redis.Client, numWorkers int, userService *service.UserService) {
	_, err := client.XGroupCreateMkStream(ctx, queue.SystemTaskNameCounterRecalc, groupName, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Failed to create consumer group: %v", err)
	}

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for {
				// Читаем сообщения из стрима
				result, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    groupName,
					Consumer: string(rune(workerID)),
					Streams:  []string{queue.SystemTaskNameCounterRecalc, ">"},
					Count:    10,
					Block:    5 * time.Second,
				}).Result()

				if err != nil && !errors.Is(err, redis.Nil) || len(result) == 0 {
					time.Sleep(1 * time.Second)
					continue
				}

				for _, msg := range result[0].Messages {
					userIDStr, ok := msg.Values["user_id"].(string)
					if !ok {
						log.Printf("Worker %d: invalid user_id type", workerID)
						continue
					}

					receiverID, err := strconv.Atoi(userIDStr)
					if err != nil {
						log.Printf("Worker %d: failed to parse user_id: %v", workerID, err)
						continue
					}

					log.Printf("Worker %d: start CountersWorkers for %d\n", workerID, receiverID)

					err = userService.UpdateDialogMessagesUnreadCounter(ctx, int64(receiverID))
					if err != nil {
						log.Printf("Worker %d: failed update counter user_id %d: %v", workerID, receiverID, err)
						continue
					}

					client.XAck(ctx, queue.SystemTaskNameCounterRecalc, groupName, msg.ID)
				}
			}
		}(i)
	}
}
