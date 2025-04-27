package service

import (
	"context"
	"errors"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport"
	"otus-highload-arh-homework/internal/social/transport/dto"
)

type userUserCase interface {
	GetUserByID(ctx context.Context, requestedID string) (*dto.UserResponse, error)
}

type UserService struct {
	userUC     userUserCase
	jwtService *JWTGenerator
}

func NewUserService(
	userUC userUserCase,
	jwtService *JWTGenerator,
) *UserService {
	return &UserService{
		userUC:     userUC,
		jwtService: jwtService,
	}
}

// GetUserByID возвращает информацию о пользователе
func (s *UserService) GetUserByID(ctx context.Context, requestedID string) (*dto.UserResponse, error) {
	// Получение пользователя
	user, err := s.userUC.GetUserByID(ctx, requestedID)
	if err != nil {
		if errors.Is(err, transport.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Преобразование в DTO
	response := dto.ConvertUserToResponse(user)
	return &response, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, userID string, input dto.UpdateUserInput, authToken string) (*dto.UserResponse, error) {
	// Валидация токена
	authUserID, err := s.jwtService.ValidateToken(authToken)
	if err != nil {
		return nil, err
	}

	// Проверка что пользователь обновляет свои данные
	if userID != authUserID {
		return nil, ErrPermissionDenied
	}

	// Преобразование DTO -> Entity
	updateData := entity.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		BirthDate: input.BirthDate,
		Gender:    input.Gender,
		Interests: input.Interests,
		City:      input.City,
	}

	// Обновление данных
	updatedUser, err := s.userUC.Update(ctx, userID, updateData)
	if err != nil {
		return nil, err
	}

	// Преобразование в DTO
	response := dto.ConvertUserToResponse(updatedUser)
	return &response, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, userID string, authToken string) error {
	// Валидация токена
	authUserID, err := s.jwtService.ValidateToken(authToken)
	if err != nil {
		return err
	}

	// Проверка что пользователь удаляет себя
	if userID != authUserID {
		return ErrPermissionDenied
	}

	return s.userUC.Delete(ctx, userID)
}
