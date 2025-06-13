package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport/dto"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
)

type userUserCase interface {
	GetByID(ctx context.Context, id int) (*entity.User, error)
	Search(ctx context.Context, firstName, lastName string) ([]*entity.User, error)
}

type friendUseCase interface {
	AddFriend(ctx context.Context, userID, friendID int) error
	RemoveFriend(ctx context.Context, userID, friendID int) error
	CheckFriendship(ctx context.Context, userID, friendID int) (bool, error)
}

type UserService struct {
	userUC   userUserCase
	friendUC friendUseCase
}

func NewUserService(
	userUC userUserCase,
	friendUC friendUseCase,
) *UserService {
	return &UserService{
		userUC:   userUC,
		friendUC: friendUC,
	}
}

// GetUserByID возвращает информацию о пользователе
func (s *UserService) GetUserByID(ctx context.Context, subID int, requestID string) (*dto.UserResponse, error) {
	if subID == 0 {
		return nil, errors.New("user not found in context")
	}

	userID, err := strconv.Atoi(requestID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id %s", requestID)
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

// SearchUsers возвращает список пользователей по имени и фамилии
func (s *UserService) SearchUsers(ctx context.Context, firstName, lastName string) ([]dto.UserResponse, error) {
	// Валидация параметров
	if len(firstName) < 2 || len(lastName) < 2 {
		return nil, errors.New("search query must be at least 2 characters long")
	}

	users, err := s.userUC.Search(ctx, firstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Преобразование в DTO
	var response []dto.UserResponse
	for _, user := range users {
		response = append(response, dto.ConvertUserToResponse(user))
	}

	return response, nil
}

// SetFriend добавляет пользователя в друзья
func (s *UserService) SetFriend(ctx context.Context, userID int, friendIDStr string) error {
	// Валидация ID друга
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFriendID, err)
	}

	// Проверка на добавление самого себя
	if userID == friendID {
		return ErrSelfOperation
	}

	// Вызов use case
	err = s.friendUC.AddFriend(ctx, userID, friendID)
	if err != nil {
		// Маппинг ошибок из UC в ошибки сервиса
		switch {
		case errors.Is(err, userUC.ErrUserNotFound):
			return ErrUserNotFound
		case errors.Is(err, userUC.ErrSelfOperation):
			return ErrSelfOperation
		case errors.Is(err, userUC.ErrAlreadyFriends):
			return ErrAlreadyFriends
		default:
			return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		}
	}

	return nil
}

// DeleteFriend удаляет пользователя из друзей
func (s *UserService) DeleteFriend(ctx context.Context, userID int, friendIDStr string) error {
	// Валидация ID друга
	friendID, err := strconv.Atoi(friendIDStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidFriendID, err)
	}

	if userID == friendID {
		return ErrSelfOperation
	}

	err = s.friendUC.RemoveFriend(ctx, userID, friendID)
	if err != nil {
		switch {
		case errors.Is(err, userUC.ErrUserNotFound):
			return ErrUserNotFound
		case errors.Is(err, userUC.ErrSelfOperation):
			return ErrSelfOperation
		case errors.Is(err, userUC.ErrNotFriends):
			return ErrNotFriends
		default:
			return fmt.Errorf("%w: %v", ErrDatabaseOperation, err)
		}
	}

	return nil
}
