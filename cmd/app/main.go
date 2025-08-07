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
	cachewarmer "otus-highload-arh-homework/internal/social/transport/cache"
	"otus-highload-arh-homework/internal/social/transport/clients/dialog/grpc"
	"otus-highload-arh-homework/internal/social/transport/server"
	authInternal "otus-highload-arh-homework/internal/social/transport/service"
	authUC "otus-highload-arh-homework/internal/social/usecase/auth"
	postUC "otus-highload-arh-homework/internal/social/usecase/post"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
	"otus-highload-arh-homework/pkg/auth"
	"otus-highload-arh-homework/pkg/clients/kafka"
	"otus-highload-arh-homework/pkg/clients/pg"
	"otus-highload-arh-homework/pkg/clients/redis"

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

	cfg := config.Load()

	// 1. Инициализация ресурсов
	pgPool, err := pg.New(ctx, &cfg.PG)
	if err != nil {
		log.Fatalf("Failed to init PG: %v", err)
	}
	defer pgPool.Close()

	if err := runMigrations(pgPool.Config().ConnString()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	redisClient, err := redis.New(ctx, &cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer func() {
		if err := redis.Close(redisClient); err != nil {
			log.Printf("Failed to close Redis connection: %v", err)
		}
	}()

	// Инициализация Kafka Producer
	feedProducer := kafka.NewProducer(
		[]string{cfg.Kafka.Address},
		cfg.Kafka.FeedUpdatesTopic,
	)
	defer func(feedProducer *kafka.Producer) {
		err := feedProducer.Close()
		if err != nil {
			log.Printf("Failed to close Kafka producer: %v", err)
		}
	}(feedProducer)

	dialogClient, err := grpc.New(cfg.Dialog.ClientAddr, cfg.Dialog.Timeout)
	if err != nil {
		log.Fatalf("Failed to initialize Dialog gRPC client: %v", err)
	}
	defer func(dialogClient *grpc.Client) {
		err = errors.Join(err, dialogClient.Close())
	}(dialogClient)

	defer func(dialogClient *grpc.Client) {
		err = errors.Join(err, dialogClient.Close())
	}(dialogClient)

	// 2. Вспомогательные сервисы
	hasher := auth.NewBcryptHasher(cfg.Auth.HashCost)

	// 3. Репозитории
	userRepo := postgres.NewUserRepository(pgPool)
	postRepo := postgres.NewPostRepository(pgPool)

	// 4. Инициализация очереди и CacheWarmer
	redisQueue := cachewarmer.NewRedisQueue(redisClient)
	cacheWarmer := cachewarmer.New(redisQueue, redisClient)

	// 5. Бизнес слои
	authUseCase := authUC.NewAuth(userRepo, hasher, cacheWarmer)
	userUseCase := userUC.New(userRepo)
	friendUseCase := userUC.NewFriendUseCase(userRepo)
	postUseCase := postUC.NewPostUseCase(postRepo)

	// 6. Сервисы транспортного уровня
	jwtService := authInternal.NewJWTGenerator(cfg.Auth.JwtSecretKey, cfg.Auth.JwtDuration)
	authService := authInternal.NewAuthService(authUseCase, jwtService)
	userService := authInternal.NewUserService(userUseCase, friendUseCase, dialogClient)
	postService := authInternal.NewPostService(postUseCase, friendUseCase, cacheWarmer, feedProducer)

	// 7. Запуск воркеров для обработки задач прогрева кэша
	go func() {
		log.Println("Starting StartCacheWorkers...", cfg.Cache.NumWorkers)
		cachewarmer.StartCacheWorkers(ctx, redisClient, cfg.Cache.NumWorkers, postService)
	}()

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
