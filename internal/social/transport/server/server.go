package server

import (
	"github.com/gin-gonic/gin"
	"otus-highload-arh-homework/internal/social/delivery/http"
	"otus-highload-arh-homework/internal/social/transport/services"
)

type Server struct {
	router      *gin.Engine
	authHandler *http.AuthHandler
}

func New(
	authService *services.AuthService,
) *Server {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Инициализация handler'ов
	authHandler := http.NewAuthHandler(authService)

	// Роуты
	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			// authGroup.POST("/login", authHandler.Login)
		}
	}

	return &Server{
		router:      router,
		authHandler: authHandler,
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
