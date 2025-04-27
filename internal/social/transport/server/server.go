package server

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"otus-highload-arh-homework/internal/social/delivery/http"
	"otus-highload-arh-homework/internal/social/transport/service"
)

type Server struct {
	router      *gin.Engine
	authHandler *http.AuthHandler
}

func New(
	authService *service.AuthService,
) *Server {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Инициализация handler'ов
	authHandler := http.NewAuthHandler(authService)

	// Swagger docs route
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
