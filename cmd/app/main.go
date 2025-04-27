package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"otus-highload-arh-homework/internal/social/config"
	"otus-highload-arh-homework/internal/social/repository/postgres"
	"otus-highload-arh-homework/internal/social/transport/server"
	authInternal "otus-highload-arh-homework/internal/social/transport/services"
	authUC "otus-highload-arh-homework/internal/social/usecase/auth"
	"otus-highload-arh-homework/pkg/auth"
	"otus-highload-arh-homework/pkg/pg"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Инициализация ресурсов
	pgPool, err := pg.New(ctx, pg.Load())
	if err != nil {
		log.Fatalf("Failed to init PG: %v", err)
	}
	defer pgPool.Close()

	cfg := config.Load()

	// 2. Вспомогательные сервисы
	hasher := auth.NewBcryptHasher(cfg.Auth.HashCost)

	// 3. Репозитории
	userRepo := postgres.NewUserRepository(pgPool)

	// 4. Бизнес слои
	authUseCase := authUC.NewAuth(userRepo, hasher)

	// 5. Сервисы транспортного уровня
	jwtService := authInternal.NewJWTGenerator(cfg.Auth.JwtSecretKey, cfg.Auth.JwtDuration)
	authService := authInternal.NewAuthService(authUseCase, jwtService)

	// Запуск сервера
	srv := server.New(authService)
	if err := srv.Run(":8080"); err != nil {
		panic(err)
	}
}
