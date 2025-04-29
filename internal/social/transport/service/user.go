package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport/dto"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
)

type userUserCase interface {
	GetByID(ctx context.Context, id int) (*entity.User, error)
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
func (s *UserService) GetUserByID(ctx context.Context, subID int, requestID string) (*dto.UserResponse, error) {
	// Валидируем запрос
	if subID == 0 {
		return nil, errors.New("user not found in context")
	}

	userID, err := strconv.Atoi(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id %s", requestID)
	}

	if userID != subID {
		log.Println("user not found in context", requestID, subID)
		return nil, fmt.Errorf("incorrect user id %s", requestID)
	}

	// Получение пользователя
	user, err := s.userUC.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userUC.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Преобразование в DTO
	response := dto.ConvertUserToResponse(user)
	return &response, nil
}
