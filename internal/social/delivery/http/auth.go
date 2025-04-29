package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "otus-highload-arh-homework/docs"
	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport"
	"otus-highload-arh-homework/internal/social/transport/dto"
	auth2 "otus-highload-arh-homework/internal/social/transport/service"
)

type AuthHandler struct {
	authService *auth2.AuthService
}

func NewAuthHandler(authService *auth2.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Регистрация нового пользователя
// @Description Создает нового пользователя и возвращает JWT токен
// @Tags auth
// @Accept  json
// @Produce  json
// @Param input body dto.RegisterInput true "Данные для регистрации"
// @Success 201 {object} dto.RegisterSuccessResponse "Успешная регистрация"
// @Failure 400 {object} dto.ValidationErrorResponse "Неверные входные данные"
// @Failure 409 {object} dto.ErrorResponse "Пользователь уже существует"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterInput

	// Парсинг и валидация входных данных
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		log.Println(err)
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
	log.Println(err)
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
			"error":   "Internal server error",
			"details": err.Error(),
		})
	}
}

// Login godoc
// @Summary Аутентификация пользователя
// @Description Возвращает JWT токен для аутентификации
// @Tags auth
// @Accept json
// @Produce json
// @Param input body dto.LoginInput true "Данные для входа"
// @Success 200 {object} dto.RegisterSuccessResponse "Успешный вход"
// @Failure 400 {object} dto.ErrorResponse "Неверные входные данные"
// @Failure 401 {object} dto.ErrorResponse "Неверный email или пароль"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginInput

	// Парсинг и валидация входных данных
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input data",
			"details": err.Error(),
		})
		return
	}

	// Вызов сервиса
	userResponse, token, err := h.authService.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"user":  userResponse,
		"token": token,
	})
}
