package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"otus-highload-arh-homework/internal/repository/postgres"
	"otus-highload-arh-homework/internal/server/grpc"
	"otus-highload-arh-homework/internal/social/config"
	postgres2 "otus-highload-arh-homework/internal/social/repository/postgres"
	"otus-highload-arh-homework/internal/usecase/dialog"
	"otus-highload-arh-homework/pkg/clients/pg"
	dialogv1 "otus-highload-arh-homework/pkg/proto/dialog/v1"

	"google.golang.org/grpc"
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

	// 3. Репозитории
	userRepo := postgres2.NewUserRepository(pgPool)

	//todoo

	// Инициализация слоев
	repo := postgres.NewDialogRepository(pgPool)
	uc := dialog.New(repo)
	service := grpc.NewDialogService(uc)

	// Запуск gRPC сервера
	srv := grpc.NewServer()
	dialogv1.RegisterDialogServiceServer(srv, service)

	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on %s", lis.Addr())
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-shutdownCtx.Done():
		srv.Stop()
	}

	log.Println("Server stopped gracefully")
}
