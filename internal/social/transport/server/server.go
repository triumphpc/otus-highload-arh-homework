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
	userHandler *http.UserHandler
	jwtService  *service.JWTGenerator
}

func New(
	authService *service.AuthService,
	userService *service.UserService,
	jwtService *service.JWTGenerator,
) *Server {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Инициализация handler'ов
	authHandler := http.NewAuthHandler(authService)
	userHandler := http.NewUserHandler(userService)

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

		userGroup := api.Group("/user")
		userGroup.Use(http.AuthMiddleware(jwtService))
		{
			userGroup.GET("/get/:id", userHandler.GetUser)
		}
	}

	return &Server{
		router:      router,
		authHandler: authHandler,
		userHandler: userHandler,
		jwtService:  jwtService,
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
