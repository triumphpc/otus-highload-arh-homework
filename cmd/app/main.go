package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	authInternal "otus-highload-arh-homework/internal/social/auth"
	"otus-highload-arh-homework/internal/social/config"
	"otus-highload-arh-homework/internal/social/usecase/repository/postgres"
	"otus-highload-arh-homework/pkg/auth"
	"otus-highload-arh-homework/pkg/pg"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Инициализация PG
	pgPool, err := pg.New(ctx, pg.Load())
	if err != nil {
		log.Fatalf("Failed to init PG: %v", err)
	}
	defer pgPool.Close()

	cfg := config.Load()

	// 2. Auth dependencies
	hasher := auth.NewBcryptHasher(cfg.Auth.HashCost)
	jwtService := authInternal.NewJWTGenerator("your-secret-key", 24*time.Hour)
	authService := auth.NewAuthService(userRepo, hasher, jwtService)

	// 2. Инициализация слоёв
	userRepo := postgres.NewUserRepository(pgPool)
	userUC := user.NewUseCase(userRepo)
	authUC := auth.NewUseCase(userRepo)

	// Handlers
	authHandler := http.NewAuthHandler(authService)
	userHandler := http.NewUserHandler(userRepo)

	//userRepo := postgres.NewUserRepo(pg)
	//userUC := user.New(userRepo)
	//authUC := auth.New(userRepo)
	//
	//// Роутер
	//r := http.NewRouter(authUC, userUC)
	//
	//// Запуск сервера
	//log.Printf("Server started on :%d", cfg.HTTP.Port)
	//http.ListenAndServe(":"+cfg.HTTP.Port, r)
}
