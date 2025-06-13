package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"otus-highload-arh-homework/internal/social/config"
	"otus-highload-arh-homework/internal/social/repository/postgres"
	"otus-highload-arh-homework/internal/social/transport/server"
	authInternal "otus-highload-arh-homework/internal/social/transport/service"
	authUC "otus-highload-arh-homework/internal/social/usecase/auth"
	postUC "otus-highload-arh-homework/internal/social/usecase/post"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
	"otus-highload-arh-homework/pkg/auth"
	"otus-highload-arh-homework/pkg/clients/pg"

	"github.com/pressly/goose/v3"
)

// @title Social Network API
// @version 1.0
// @description API для социальной сети

// @contact.name API Support
// @contact.email trumph.job@gmail.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http

func main() {
	log.Println("Starting application...")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Инициализация ресурсов
	pgPool, err := pg.New(ctx, pg.Load())
	if err != nil {
		log.Fatalf("Failed to init PG: %v", err)
	}
	defer pgPool.Close()

	if err := runMigrations(pgPool.Config().ConnString()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	cfg := config.Load()

	// 2. Вспомогательные сервисы
	hasher := auth.NewBcryptHasher(cfg.Auth.HashCost)

	// 3. Репозитории
	userRepo := postgres.NewUserRepository(pgPool)
	postRepo := postgres.NewPostRepository(pgPool)

	// 4. Бизнес слои
	authUseCase := authUC.NewAuth(userRepo, hasher)
	userUseCase := userUC.New(userRepo)
	friendUseCase := userUC.NewFriendUseCase(userRepo)
	postUseCase := postUC.NewPostUseCase(postRepo)

	// 5. Сервисы транспортного уровня
	jwtService := authInternal.NewJWTGenerator(cfg.Auth.JwtSecretKey, cfg.Auth.JwtDuration)
	authService := authInternal.NewAuthService(authUseCase, jwtService)
	userService := authInternal.NewUserService(userUseCase, friendUseCase)
	postService := authInternal.NewPostService(postUseCase)

	srv := server.New(authService, userService, postService, jwtService)

	// Запуск сервера
	go func() {
		if err := srv.Run(cfg.HTTP.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func runMigrations(dbURL string) error {
	// Установка соединения для goose
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err = errors.Join(err, db.Close())
	}(db)

	// Настройка goose
	goose.SetBaseFS(os.DirFS("migrations"))

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Применение миграций
	return goose.Up(db, ".")
}
