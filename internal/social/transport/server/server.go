package server

import (
	"context"
	"errors"
	ht "net/http"

	"otus-highload-arh-homework/internal/social/handler/http"
	"otus-highload-arh-homework/internal/social/transport/service"
	"otus-highload-arh-homework/internal/social/transport/websocket"
	"otus-highload-arh-homework/pkg/clients/prometheus"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	router      *gin.Engine
	authHandler *http.AuthHandler
	userHandler *http.UserHandler
	postHandler *http.PostHandler
	jwtService  *service.JWTGenerator
	httpServer  *ht.Server
	wsServer    *websocket.Server
}

func New(
	authService *service.AuthService,
	userService *service.UserService,
	postService *service.PostService,
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
	postHandler := http.NewPostHandler(postService)

	// Инициализация WebSocket сервера
	//wsServer := websocket.NewServer()

	// Swagger docs route
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Регистрируем метрики
	router.Use(prometheus.MetricsMiddleware())
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", healthCheck)

	// Роуты
	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		dialogGroup := api.Group("/dialog")
		dialogGroup.Use(http.AuthMiddleware(jwtService))
		{
			dialogGroup.POST("/:user_id/send", userHandler.SendDialogMessage)
			dialogGroup.GET("/:user_id/list", userHandler.GetDialogMessages)
		}

		userGroup := api.Group("/user")
		userGroup.Use(http.AuthMiddleware(jwtService))
		{
			userGroup.GET("/get/:id", userHandler.GetUser)
			userGroup.GET("/search", userHandler.SearchUsers)
		}
	}

	// Друзья
	friendGroup := api.Group("/friend")
	friendGroup.Use(http.AuthMiddleware(jwtService))
	{
		friendGroup.PUT("/set/:user_id", userHandler.SetFriend)
		friendGroup.PUT("/delete/:user_id", userHandler.DeleteFriend)
	}

	// Посты
	postGroup := api.Group("/post")
	postGroup.Use(http.AuthMiddleware(jwtService))
	{
		postGroup.POST("/create", postHandler.CreatePost)
		postGroup.PUT("/update", postHandler.UpdatePost)
		postGroup.PUT("/delete/:id", postHandler.DeletePost)
		postGroup.GET("/get/:id", postHandler.GetPost)
		postGroup.GET("/feed", postHandler.GetFeed)
	}

	// Новая версия API (v2)
	v2 := router.Group("/api/v2")
	{
		dialogGroup := v2.Group("/dialog")
		dialogGroup.Use(http.AuthMiddleware(jwtService))
		{
			dialogGroup.POST("/:user_id/send", userHandler.SendDialogMessageV2)
			dialogGroup.GET("/:user_id/list", userHandler.GetDialogMessagesV2)
		}
	}

	return &Server{
		router:      router,
		authHandler: authHandler,
		userHandler: userHandler,
		jwtService:  jwtService,
	}
}

// Run запускает сервер с поддержкой graceful shutdown
func (s *Server) Run(addr string) error {
	s.httpServer = &ht.Server{
		Addr:    addr,
		Handler: s.router,
	}

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, ht.ErrServerClosed) {
		return err
	}

	return nil
}

// Shutdown корректно останавливает сервер с таймаутом
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func healthCheck(c *gin.Context) {
	c.JSON(ht.StatusOK, gin.H{
		"status": "OK",
	})
}
