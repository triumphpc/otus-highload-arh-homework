package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"otus-highload-arh-homework/internal/social/config"
	postgres2 "otus-highload-arh-homework/internal/social/repository/postgres"
	cachewarmer "otus-highload-arh-homework/internal/social/transport/cache"
	"otus-highload-arh-homework/internal/social/transport/server"
	authInternal "otus-highload-arh-homework/internal/social/transport/service"
	authUC "otus-highload-arh-homework/internal/social/usecase/auth"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
	"otus-highload-arh-homework/pkg/auth"
	"otus-highload-arh-homework/pkg/clients/pg"
	"otus-highload-arh-homework/pkg/clients/redis"
	"otus-highload-arh-homework/pkg/queue"
)

func main() {
	log.Println("Starting application...")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	// 1. Инициализация ресурсов
	pgPool, err := pg.New(ctx, &cfg.PG)
	if err != nil {
		log.Fatalf("Failed to init PG: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redis.Close(redisClient); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	// Вспомогательные
	hasher := auth.NewBcryptHasher(cfg.Auth.HashCost)

	// 3. Репозитории
	userRepo := postgres2.NewUserRepository(pgPool)

	// Очереди
	redisTaskQueue := queue.NewRedisQueue(redisClient)
	redisQueue := cachewarmer.NewRedisQueue(redisClient)
	cacheWarmer := cachewarmer.New(redisQueue, redisClient)

	// Бизнес слой
	authUseCase := authUC.NewAuth(userRepo, hasher, cacheWarmer)
	// todo
	userUseCase := userUC.New(userRepo, redisTaskQueue, cacheWarmer)

	// Транспортный уровень
	jwtService := authInternal.NewJWTGenerator(cfg.Auth.JwtSecretKey, cfg.Auth.JwtDuration)
	authService := authInternal.NewAuthService(authUseCase, jwtService)
	userService := authInternal.NewUserService(userUseCase, nil, nil)

	srv := server.NewCounterServer(authService, userService, jwtService)

	// Запуск сервера
	go func() {
		if err := srv.RunCounterServer(cfg.HTTPCounter.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("Server started. Press Ctrl+C to stop.")

	// Ожидаем сигнал завершения
	<-ctx.Done()

	// Graceful shutdown (даём 5 секунд на завершение операций)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}
