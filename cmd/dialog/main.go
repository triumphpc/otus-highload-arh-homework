package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"otus-highload-arh-homework/internal/social/config"
	postgres2 "otus-highload-arh-homework/internal/social/repository/postgres"
	grpcServer "otus-highload-arh-homework/internal/social/transport/server/dialog/grpc"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
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

	// 3. Репозитории
	userRepo := postgres2.NewUserRepository(pgPool)
	redisTaskQueue := queue.NewRedisQueue(redisClient)
	userUseCase := userUC.New(userRepo, redisTaskQueue)

	srv, err := grpcServer.New(userUseCase, cfg.Dialog.Address)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on port %s", cfg.Dialog.Address)
		if err := srv.Run(); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Println("Shutting down server...")

	srv.Stop()
	log.Println("Server stopped gracefully")
}
