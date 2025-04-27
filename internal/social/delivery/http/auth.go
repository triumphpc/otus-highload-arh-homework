package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport"
	"otus-highload-arh-homework/internal/social/transport/dto"
	auth2 "otus-highload-arh-homework/internal/social/transport/services"
)

type AuthHandler struct {
	authService *auth2.AuthService
}

func NewAuthHandler(authService *auth2.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterInput

	// Парсинг и валидация входных данных
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Вызов сервиса
	userResponse, token, err := h.authService.Register(c.Request.Context(), input)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Успешный ответ
	c.JSON(http.StatusCreated, gin.H{
		"user":  userResponse,
		"token": token,
	})
}

func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, transport.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already in use",
		})
	case errors.Is(err, entity.ErrUnderageUser):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User must be at least 18 years old",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}
