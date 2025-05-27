package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"otus-highload-arh-homework/internal/social/transport/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetUser godoc
// @Summary Пользователь по ID
// @Description Получить информацию по пользователю
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/get/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	subID := c.Value("userID").(int)

	user, err := h.userService.GetUserByID(c.Request.Context(), subID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if errors.Is(err, service.ErrPermissionDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
			return
		}

		log.Println(fmt.Errorf("GetUserByID: %w", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// SearchUsers godoc
// @Summary Поиск пользователей
// @Description Поиск анкет пользователей по имени и фамилии
// @Tags user
// @Accept json
// @Produce json
// @Param first_name query string true "Часть имени для поиска"
// @Param last_name query string true "Часть фамилии для поиска"
// @Security ApiKeyAuth
// @Success 200 {array} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /user/search [get]
func (h *UserHandler) SearchUsers(c *gin.Context) {
	firstName := c.Query("first_name")
	lastName := c.Query("last_name")

	users, err := h.userService.SearchUsers(c.Request.Context(), firstName, lastName)
	if err != nil {
		log.Println(fmt.Errorf("SearchUsers: %w", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
