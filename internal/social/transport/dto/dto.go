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
	Error string `json:"error"`
}
