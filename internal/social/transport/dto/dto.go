package dto

import (
	"time"

	"otus-highload-arh-homework/internal/social/entity"
)

type RegisterInput struct {
	FirstName string        `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string        `json:"last_name" validate:"required,min=2,max=50"`
	Email     string        `json:"email" validate:"required,email"`
	Password  string        `json:"password" validate:"required,min=8,max=72"`
	BirthDate time.Time     `json:"birth_date" validate:"required" example:"1983-01-02T15:04:05Z" format:"date-time"`
	Gender    entity.Gender `json:"gender" validate:"required,oneof=male female other"`
	Interests []string      `json:"interests" validate:"max=10"`
	City      string        `json:"city" validate:"required,max=100"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        int           `json:"id"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	BirthDate time.Time     `json:"birth_date"`
	Gender    entity.Gender `json:"gender"`
	Interests []string      `json:"interests"`
	City      string        `json:"city"`
	CreatedAt time.Time     `json:"created_at"`
	IsAdult   bool          `json:"is_adult"`
}

// Успешный ответ
type RegisterSuccessResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// Ошибка валидации
type ValidationErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

// Внутренняя ошибка сервера
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

type CreatePostRequest struct {
	Text string `json:"text" binding:"required"`
}

type UpdatePostRequest struct {
	PostID string `json:"id" binding:"required"`
	Text   string `json:"text" binding:"required"`
}

type PostIdResponse struct {
	PostID string `json:"id"`
}

type PostResponse struct {
	ID        string    `json:"id"`
	AuthorID  int       `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FeedUpdateEvent struct {
	PostID    string `json:"post_id"`
	AuthorID  int    `json:"author_id"`
	Action    string `json:"action"` // "create", "update", "delete"
	Text      string `json:"text,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

type SendMessageRequest struct {
	Text string `json:"text" binding:"required,min=1,max=1000"`
}

type DialogMessage struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Text       string    `json:"text"`
	SentAt     time.Time `json:"-"`
	SentAtStr  string    `json:"sent_at"`
	IsOwn      bool      `json:"is_own"`
}

// SuccessResponse - универсальный DTO для успешных ответов
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DialogMessageV2 struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Text       string    `json:"text"`
	SentAt     time.Time `json:"sent_at"`
	IsOwn      bool      `json:"is_own"`
}

type ErrorResponseV2 struct {
	Error     string    `json:"error"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

type SuccessResponseV2 struct {
	Status    string      `json:"status"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}
