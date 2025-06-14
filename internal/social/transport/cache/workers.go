package cachewarmer

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"

	"otus-highload-arh-homework/internal/social/transport/service"

	"github.com/redis/go-redis/v9"
)

func StartCacheWorkers(ctx context.Context, client *redis.Client, numWorkers int, postService *service.PostService) {
	_, err := client.XGroupCreateMkStream(ctx, "cache_warm_tasks", "cache_workers", "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("Failed to create consumer group: %v", err)
	}

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			for {
				// Читаем сообщения из стрима
				result, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    "cache_workers",
					Consumer: string(rune(workerID)),
					Streams:  []string{"cache_warm_tasks", ">"},
					Count:    10,
					Block:    5 * time.Second,
				}).Result()

				if err != nil && !errors.Is(err, redis.Nil) {
					log.Printf("Worker %d error: %v", workerID, err)
					time.Sleep(1 * time.Second)
					continue
				}

				for _, msg := range result[0].Messages {
					userIDStr, ok := msg.Values["user_id"].(string)
					if !ok {
						log.Printf("Worker %d: invalid user_id type", workerID)
						continue
					}

					authorID, err := strconv.Atoi(userIDStr)
					if err != nil {
						log.Printf("Worker %d: failed to parse user_id: %v", workerID, err)
						continue
					}

					log.Printf("Worker %d: start PreloadUserFriendsFeeds for %d\n", workerID, authorID)

					if err := postService.PreloadUserFriendsFeeds(ctx, authorID); err != nil {
						log.Printf("Worker %d failed to preload feed for author %d: %v",
							workerID, authorID, err)
					}

					client.XAck(ctx, "cache_warm_tasks", "cache_workers", msg.ID)
				}
			}
		}(i)
	}
}
