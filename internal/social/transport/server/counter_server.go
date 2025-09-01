package server

import (
	"context"
	"errors"
	ht "net/http"

	"otus-highload-arh-homework/internal/social/handler/http"
	"otus-highload-arh-homework/internal/social/transport/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type CounterServer struct {
	router      *gin.Engine
	authHandler *http.AuthHandler
	userHandler *http.UserHandler
	jwtService  *service.JWTGenerator
	httpServer  *ht.Server
}

func NewCounterServer(
	authService *service.AuthService,
	userService *service.UserService,
	jwtService *service.JWTGenerator,
) *Server {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(RequestIDMiddleware())

	// Инициализация handler'ов
	authHandler := http.NewAuthHandler(authService)
	userHandler := http.NewUserHandler(userService)

	// Swagger docs route
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", healthCheckCounterServer)

	// Роуты
	api := router.Group("/api/v1")
	{
		dialogGroup := api.Group("/dialog")
		dialogGroup.Use(http.AuthMiddleware(jwtService))
		{
			dialogGroup.GET("/count", userHandler.GetUnreadDialogMessagesCount)
		}
	}

	return &Server{
		router:      router,
		authHandler: authHandler,
		userHandler: userHandler,
		jwtService:  jwtService,
	}
}

// RunCounterServer запускает сервер с поддержкой graceful shutdown
func (s *Server) RunCounterServer(addr string) error {
	s.httpServer = &ht.Server{
		Addr:    addr,
		Handler: s.router,
	}

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, ht.ErrServerClosed) {
		return err
	}

	return nil
}

// ShutdownCounterServer корректно останавливает сервер с таймаутом
func (s *Server) ShutdownCounterServer(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func healthCheckCounterServer(c *gin.Context) {
	c.JSON(ht.StatusOK, gin.H{
		"status": "OK",
	})
}
