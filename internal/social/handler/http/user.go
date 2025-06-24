package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"otus-highload-arh-homework/internal/social/repository"
	"otus-highload-arh-homework/internal/social/transport/dto"
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

// SendDialogMessage godoc
// @Summary Отправить сообщение пользователю
// @Tags user
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя-получателя"
// @Param input body dto.SendMessageRequest true "Текст сообщения"
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponse
// @Router /dialog/{user_id}/send [post]
func (h *UserHandler) SendDialogMessage(c *gin.Context) {
	receiverIDStr := c.Param("user_id")

	receiverID, err := strconv.ParseInt(receiverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid recipient ID format",
			Details: "User ID must be a number",
		})
		return
	}

	senderID := c.Value("userID").(int)
	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.Text) == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Message text cannot be empty",
		})
		return
	}

	err = h.userService.SendDialogMessage(
		c.Request.Context(),
		int64(senderID),
		receiverID,
		req.Text,
	)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "Recipient not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to send message",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Status:  "success",
		Message: "Message sent successfully",
	})
}

// GetDialogMessages godoc
// @Summary Получить диалог с пользователем
// @Tags user
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Security ApiKeyAuth
// @Success 200 {array} dto.DialogMessage
// @Router /dialog/:user_id/list [get]
func (h *UserHandler) GetDialogMessages(c *gin.Context) {
	otherUserIDStr := c.Param("user_id")

	otherUserID, err := strconv.ParseInt(otherUserIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "Invalid user ID format",
			Details: "User ID must be a numeric value",
		})
		return
	}

	currentUserID := c.Value("userID").(int)

	messages, err := h.userService.GetDialogMessages(
		c.Request.Context(),
		int64(currentUserID),
		otherUserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: "User not found",
			})
		case errors.Is(err, repository.ErrNoMessagesFound):
			c.JSON(http.StatusOK, []dto.DialogMessage{}) // Пустой массив вместо ошибки
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "Failed to get dialog messages",
				Details: err.Error(),
			})
		}

		return
	}

	c.JSON(http.StatusOK, messages)
}
