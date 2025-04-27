package dto

import (
	"time"

	"otus-highload-arh-homework/internal/social/entity"
)

// ConvertRegisterInputToUser преобразует RegisterInput в entity.User
func ConvertRegisterInputToUser(input RegisterInput) *entity.User {
	return &entity.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		BirthDate: input.BirthDate,
		Gender:    input.Gender,
		Interests: input.Interests,
		City:      input.City,
		CreatedAt: time.Now(), // Устанавливаем текущее время
	}
}

// ConvertUserToResponse преобразует entity.User в UserResponse
func ConvertUserToResponse(user *entity.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		BirthDate: user.BirthDate,
		Gender:    user.Gender,
		Interests: user.Interests,
		City:      user.City,
		CreatedAt: user.CreatedAt,
		IsAdult:   user.IsAdult(),
	}
}

// ConvertLoginInputToCredentials возвращает email и пароль из LoginInput
func ConvertLoginInputToCredentials(input LoginInput) (email, password string) {
	return input.Email, input.Password
}
