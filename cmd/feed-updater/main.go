package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"otus-highload-arh-homework/internal/social/config"
	"otus-highload-arh-homework/internal/social/handler/http"
	"otus-highload-arh-homework/internal/social/repository/postgres"
	"otus-highload-arh-homework/internal/social/transport/dto"
	"otus-highload-arh-homework/internal/social/transport/service"
	"otus-highload-arh-homework/internal/social/transport/websocket"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
	"otus-highload-arh-homework/pkg/clients/pg"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	// Инициализация Kafka Consumer
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{cfg.Kafka.Address},
		GroupID:  "feed-updaters",
		Topic:    cfg.Kafka.FeedUpdatesTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	pgPool, err := pg.New(ctx, &cfg.PG)
	if err != nil {
		log.Fatalf("Failed to init PG: %v", err)
	}
	defer pgPool.Close()

	jwtService := service.NewJWTGenerator(cfg.Auth.JwtSecretKey, cfg.Auth.JwtDuration)
	userRepo := postgres.NewUserRepository(pgPool)
	friendUseCase := userUC.NewFriendUseCase(userRepo)

	// Инициализация WebSocket сервера
	wsServer := websocket.NewServer()
	router := gin.Default()

	// Добавляем аутентификацию (используем тот же middleware, что и в API)
	router.GET("/ws/post/feed/posted", http.AuthMiddleware(jwtService), func(c *gin.Context) {
		userID := c.GetInt("userID") // Получаем из middleware
		wsServer.HandleConnection(c.Writer, c.Request, userID)
	})

	// Запускаем HTTP-сервер для WebSocket
	go func() {
		log.Printf("WebSocket server starting on %s", cfg.WS.Port)
		if err := router.Run(cfg.WS.Port); err != nil {
			log.Fatalf("WebSocket server failed: %v", err)
		}
	}()

	log.Println("Starting feed updater with WebSocket notifications...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down feed updater...")
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				log.Printf("failed to read message: %v", err)
				continue
			}

			var event dto.FeedUpdateEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("failed to unmarshal event: %v", err)
				continue
			}

			// Обработка события
			if err := processFeedUpdate(ctx, event, wsServer, friendUseCase); err != nil {
				log.Printf("failed to process feed update: %v", err)
			}
		}
	}
}

func processFeedUpdate(ctx context.Context, event dto.FeedUpdateEvent, wsServer *websocket.Server, friendUseCase *userUC.FriendUseCase) error {
	friendIDs, err := friendUseCase.GetFriendsIDs(ctx, event.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to get friends: %w", err)
	}

	// Отправляем уведомление каждому другу через WebSocket
	for _, friendID := range friendIDs {
		message := map[string]interface{}{
			"postId":         event.PostID,
			"postText":       event.Text,
			"author_user_id": event.AuthorID,
			"created_at":     event.Timestamp,
		}

		if err := wsServer.BroadcastToUser(friendID, message); err != nil {
			log.Printf("failed to send WebSocket notification to user %d: %v", friendID, err)
			// Можно добавить retry логику или dead letter queue
		} else {
			log.Printf("successfully sent notification to user %d about post %s", friendID, event.PostID)
		}
	}

	return nil
}
