package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"otus-highload-arh-homework/internal/social/transport/service"

	"github.com/gin-gonic/gin"
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

// SetFriend godoc
// @Summary Добавить в друзья
// @Description Добавить пользователя в друзья
// @Tags friend
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя, которого добавляем в друзья"
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /friend/set/{user_id} [put]
func (h *UserHandler) SetFriend(c *gin.Context) {
	friendID := c.Param("user_id")
	currentUserID := c.Value("userID").(int)

	err := h.userService.SetFriend(c.Request.Context(), currentUserID, friendID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, service.ErrSelfOperation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add yourself as friend"})
		case errors.Is(err, service.ErrAlreadyFriends):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Users are already friends"})
		case errors.Is(err, service.ErrInvalidFriendID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID format"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("SetFriend database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add friend"})
		default:
			log.Printf("SetFriend unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend added successfully"})
}

// DeleteFriend godoc
// @Summary Удалить из друзей
// @Description Удалить пользователя из друзей
// @Tags friend
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя, которого удаляем из друзей"
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /friend/delete/{user_id} [put]
func (h *UserHandler) DeleteFriend(c *gin.Context) {
	friendID := c.Param("user_id")
	currentUserID := c.Value("userID").(int)

	err := h.userService.DeleteFriend(c.Request.Context(), currentUserID, friendID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case errors.Is(err, service.ErrSelfOperation):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove yourself from friends"})
		case errors.Is(err, service.ErrNotFriends):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Users are not friends"})
		case errors.Is(err, service.ErrInvalidFriendID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend ID format"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("DeleteFriend database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove friend"})
		default:
			log.Printf("DeleteFriend unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend removed successfully"})
}
