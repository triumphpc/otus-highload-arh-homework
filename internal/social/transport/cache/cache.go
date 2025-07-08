package cachewarmer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const hasEmailPref = "has_email_"

type MessageQueue interface {
	Push(context.Context, WarmTask) error
}

type CacheWarmer struct {
	queue  MessageQueue
	client *redis.Client
	prefix string
}

func New(queue MessageQueue, client *redis.Client) *CacheWarmer {
	return &CacheWarmer{
		queue:  queue,
		client: client,
		prefix: "warm",
	}
}

func (w *CacheWarmer) WarmForNewPost(ctx context.Context, authorID int) error {
	task := WarmTask{
		UserID: authorID,
	}

	if err := w.queue.Push(ctx, task); err != nil {
		return fmt.Errorf("failed to push new post: %w", err)
	}

	log.Printf("Pushed task for author %d\n", authorID)

	return nil
}

// Set сохраняет данные в кэш с автоматической JSON сериализацией
func (w *CacheWarmer) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	fullKey := w.prefix + key

	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: failed to marshal value: %v", ErrInvalidValue, err)
	}

	if err := w.client.Set(ctx, fullKey, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("redis set operation failed: %w", err)
	}

	return nil
}

// Get получает данные из кэша с автоматической JSON десериализацией
func (w *CacheWarmer) Get(ctx context.Context, key string, dest any) error {
	fullKey := w.prefix + key

	data, err := w.client.Get(ctx, fullKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("redis get operation failed: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("%w: failed to unmarshal cached data: %v", ErrInvalidValue, err)
	}

	return nil
}

// HasEmail проверяет и сохраняет email атомарно
func (w *CacheWarmer) HasEmail(ctx context.Context, email string) (bool, error) {
	script := `
		local email_key = KEYS[1]
		
		-- Проверяем существование email
		if redis.call("EXISTS", email_key) == 1 then
			return {1}
		end
		
		-- Сохраняем email, если его нет
		redis.call("SET", email_key, 1)
		return {0}
	`

	key := hasEmailPref + email
	result, err := w.client.Eval(ctx, script, []string{key}).Result()
	if err != nil {
		return false, fmt.Errorf("redis script failed: %w", err)
	}

	// Обработка результата
	if res, ok := result.([]any); ok && len(res) == 1 {
		if exists, ok := res[0].(int64); ok {
			return exists == 1, nil
		}
	}

	return false, errors.New("unexpected response from redis")
}

// DeleteEmail удаляет email из кэша
func (w *CacheWarmer) DeleteEmail(ctx context.Context, email string) error {
	key := w.prefix + email
	if err := w.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete email: %w", err)
	}
	return nil
}
