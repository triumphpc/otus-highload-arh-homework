package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// New создает и возвращает новый Redis клиент с проверкой подключения
func New(ctx context.Context, cfg *Config) (*redis.Client, error) {
	if cfg == nil {
		return nil, errors.New("redis config is nil")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Проверяем подключение с контекстом и таймаутом
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := client.Ping(pingCtx).Result(); err != nil {
		errClose := client.Close()
		if errClose != nil {
			return nil, errors.Join(err, errClose)
		}

		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

// Close корректно закрывает соединение с Redis
func Close(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
