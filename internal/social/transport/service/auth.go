package service

import (
	"context"

	"otus-highload-arh-homework/internal/social/transport/dto"
	"otus-highload-arh-homework/internal/social/usecase/auth"
)

type AuthService struct {
	jwtService *JWTGenerator
	uc         *auth.AuthUseCase
}

// NewAuthService создает новый экземпляр AuthService
func NewAuthService(
	useCase *auth.AuthUseCase,
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

//
//// Login аутентифицирует пользователя
//func (s *AuthService) Login(ctx context.Context, email, password string) (*entity.User, string, error) {
//	// Находим пользователя
//	user, err := s.userRepo.GetByEmail(ctx, email)
//	if err != nil {
//		if errors.Is(err, repository.ErrUserNotFound) {
//			return nil, "", ErrUserNotFound
//		}
//		return nil, "", err
//	}
//
//	// Проверяем пароль
//	if !s.hasher.Check(password, user.PasswordHash) {
//		return nil, "", ErrInvalidCredentials
//	}
//
//	// Генерируем токен
//	token, err := s.jwtService.GenerateToken(user.ID)
//	if err != nil {
//		return nil, "", err
//	}
//
//	return user, token, nil
//}
//
//// VerifyToken проверяет JWT токен
//func (s *AuthService) VerifyToken(tokenStr string) (int, error) {
//	return s.jwtService.ValidateToken(tokenStr)
//}
//
//func (s *AuthService) GetUserProfile(ctx context.Context, userID int) (dto2.UserResponse, error) {
//	user, err := s.userRepo.GetByID(ctx, userID)
//	if err != nil {
//		return dto2.UserResponse{}, err
//	}
//
//	return dto2.ConvertUserToResponse(user), nil
//}
