package service

import (
	"context"
	"fmt"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport/dto"
)

type authUserCase interface {
	Register(ctx context.Context, user *entity.User, password string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
}

type AuthService struct {
	jwtService *JWTGenerator
	uc         authUserCase
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(
	useCase authUserCase,
	jwtService *JWTGenerator,
) *AuthService {
	return &AuthService{
		jwtService: jwtService,
		uc:         useCase,
	}
}

// Register регистрирует нового пользователя
func (s *AuthService) Register(ctx context.Context, input dto.RegisterInput) (*dto.UserResponse, string, error) {
	// 1. Преобразование DTO -> Entity
	user := dto.ConvertRegisterInputToUser(input)

	ok, err := user.IsValid()
	if !ok || err != nil {
		return nil, "", fmt.Errorf("invalid user %w", err)
	}

	// 2. Вызов бизнес-логики
	createdUser, err := s.uc.Register(ctx, user, input.Password)
	if err != nil {
		return nil, "", err
	}

	// 3. Генерация токена (техническая деталь)
	token, err := s.jwtService.GenerateToken(createdUser.ID)
	if err != nil {
		return nil, "", err
	}

	// 4. Преобразование в ответ
	response := dto.ConvertUserToResponse(createdUser)

	return &response, token, nil
}

// Login аутентифицирует пользователя
func (s *AuthService) Login(ctx context.Context, email, password string) (*dto.UserResponse, string, error) {
	// 1. Вызов бизнес-логики
	user, err := s.uc.Login(ctx, email, password)
	if err != nil {
		return nil, "", err
	}

	// 2. Генерация токена
	token, err := s.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	// 3. Преобразование в ответ
	response := dto.ConvertUserToResponse(user)

	return &response, token, nil
}
