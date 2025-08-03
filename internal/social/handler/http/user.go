package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"otus-highload-arh-homework/internal/social/repository"
	"otus-highload-arh-homework/internal/social/transport/dto"
	"otus-highload-arh-homework/internal/social/transport/service"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
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

// SendDialogMessageV2 godoc
// @Summary Отправить сообщение (v2)
// @Tags dialog-v2
// @Accept json
// @Produce json
// @Param user_id path string true "ID получателя"
// @Param input body dto.SendMessageRequest true "Текст сообщения"
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponseV2
// @Header 200 {string} x-request-id "Идентификатор запроса"
// @Router /api/v2/dialog/{user_id}/send [post]
func (h *UserHandler) SendDialogMessageV2(c *gin.Context) {
	requestID := c.GetString("x-request-id")
	receiverIDStr := c.Param("user_id")
	senderID := c.MustGet("userID").(int)

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseV2{
			Error:     "Invalid request body",
			RequestID: requestID,
			Timestamp: time.Now().UTC(),
		})
		return
	}

	// Вызов сервиса
	err := h.userService.SendDialogMessageV2(
		metadata.NewOutgoingContext(c.Request.Context(), metadata.Pairs("x-request-id", requestID)),
		senderID,
		receiverIDStr,
		req.Text,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseV2{
			Error:     "Failed to send message",
			RequestID: requestID,
			Timestamp: time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponseV2{
		Status:    "success",
		Message:   "Message sent",
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
	})
}

// GetDialogMessagesV2 godoc
// @Summary Получить диалог (v2)
// @Tags dialog-v2
// @Produce json
// @Param user_id path string true "ID собеседника"
// @Security ApiKeyAuth
// @Success 200 {array} dto.DialogMessageV2
// @Header 200 {string} x-request-id "Идентификатор запроса"
// @Router /api/v2/dialog/{user_id}/list [get]
func (h *UserHandler) GetDialogMessagesV2(c *gin.Context) {
	requestID := c.GetString("x-request-id")
	otherUserIDStr := c.Param("user_id")
	currentUserID := c.MustGet("userID").(int)

	// Вызов сервиса
	messages, err := h.userService.GetDialogMessagesV2(
		metadata.NewOutgoingContext(c.Request.Context(), metadata.Pairs("x-request-id", requestID)),
		currentUserID,
		otherUserIDStr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseV2{
			Error:     "Failed to get messages",
			RequestID: requestID,
			Timestamp: time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, messages)
}
