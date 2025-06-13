package http

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"otus-highload-arh-homework/internal/social/transport/dto"
	"otus-highload-arh-homework/internal/social/transport/service"

	"github.com/gin-gonic/gin"
)

// PostHandler структура для обработчиков постов
type PostHandler struct {
	postService *service.PostService
}

// NewPostHandler создает новый экземпляр PostHandler
func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

// CreatePost godoc
// @Summary Создать пост
// @Description Создает новый пост пользователя
// @Tags post
// @Accept json
// @Produce json
// @Param input body dto.CreatePostRequest true "Данные для создания поста"
// @Security ApiKeyAuth
// @Success 200 {object} dto.PostIdResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /post/create [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	currentUserID := c.Value("userID").(int)
	var input dto.CreatePostRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	postID, err := h.postService.CreatePost(c.Request.Context(), currentUserID, input.Text)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyPostText):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post text cannot be empty"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("CreatePost database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		default:
			log.Printf("CreatePost unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.PostIdResponse{PostID: postID})
}

// UpdatePost godoc
// @Summary Обновить пост
// @Description Обновляет текст существующего поста
// @Tags post
// @Accept json
// @Produce json
// @Param input body dto.UpdatePostRequest true "Данные для обновления поста"
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /post/update [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	currentUserID := c.Value("userID").(int)
	var input dto.UpdatePostRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.postService.UpdatePost(c.Request.Context(), currentUserID, input.PostID, input.Text)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		case errors.Is(err, service.ErrNotPostOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this post"})
		case errors.Is(err, service.ErrEmptyPostText):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Post text cannot be empty"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("UpdatePost database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		default:
			log.Printf("UpdatePost unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully"})
}

// DeletePost godoc
// @Summary Удалить пост
// @Description Удаляет пост по его идентификатору
// @Tags post
// @Accept json
// @Produce json
// @Param id path string true "ID поста"
// @Security ApiKeyAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /post/delete/{id} [put]
func (h *PostHandler) DeletePost(c *gin.Context) {
	currentUserID := c.Value("userID").(int)
	postID := c.Param("id")

	err := h.postService.DeletePost(c.Request.Context(), currentUserID, postID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		case errors.Is(err, service.ErrNotPostOwner):
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the owner of this post"})
		case errors.Is(err, service.ErrInvalidPostID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("DeletePost database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		default:
			log.Printf("DeletePost unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// GetPost godoc
// @Summary Получить пост
// @Description Возвращает пост по его идентификатору
// @Tags post
// @Accept json
// @Produce json
// @Param id path string true "ID поста"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /post/get/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	postID := c.Param("id")

	post, err := h.postService.GetPost(c.Request.Context(), postID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		case errors.Is(err, service.ErrInvalidPostID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID format"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("GetPost database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post"})
		default:
			log.Printf("GetPost unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetFeed godoc
// @Summary Лента постов друзей
// @Description Возвращает ленту постов друзей пользователя
// @Tags post
// @Accept json
// @Produce json
// @Param offset query int false "Оффсет" default(0)
// @Param limit query int false "Лимит" default(10)
// @Security ApiKeyAuth
// @Success 200 {array} dto.PostResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /post/feed [get]
func (h *PostHandler) GetFeed(c *gin.Context) {
	currentUserID := c.Value("userID").(int)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	posts, err := h.postService.GetFeed(c.Request.Context(), currentUserID, offset, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPaginationParams):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		case errors.Is(err, service.ErrDatabaseOperation):
			log.Printf("GetFeed database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get feed"})
		default:
			log.Printf("GetFeed unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, posts)
}
