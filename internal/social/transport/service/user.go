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
	SendDialogMessage(ctx context.Context, senderID, receiverID int64, text string) error
	GetDialogMessages(ctx context.Context, user1ID, user2ID int64) ([]*entity.DialogMessage, error)
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

	//if userID != subID {
	//	log.Println("user not found in context", requestID, subID)
	//	return nil, fmt.Errorf("incorrect user id %s", requestID)
	//}

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

func (s *UserService) SendDialogMessage(ctx context.Context, senderID, receiverID int64, text string) error {
	// Валидация
	if len(text) == 0 {
		return errors.New("message text cannot be empty")
	}

	// Проверка, что получатель существует
	exists, err := s.userUC.GetByID(ctx, int(receiverID))
	if err != nil || exists == nil {
		return errors.New("receiver not found")
	}

	return s.userUC.SendDialogMessage(ctx, senderID, receiverID, text)
}

func (s *UserService) GetDialogMessages(ctx context.Context, currentUserID, otherUserID int64) ([]dto.DialogMessage, error) {
	// Получаем сообщения из репозитория
	messages, err := s.userUC.GetDialogMessages(ctx, currentUserID, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dialog: %w", err)
	}

	// Конвертируем в DTO и устанавливаем флаг IsOwn
	currentUserIDStr := strconv.FormatInt(currentUserID, 10)
	result := make([]dto.DialogMessage, 0, len(messages))

	for _, msg := range messages {
		result = append(result, dto.DialogMessage{
			SenderID:   msg.SenderID,
			ReceiverID: msg.ReceiverID,
			Text:       msg.Text,
			SentAtStr:  msg.SentAt.Format("2006-01-02 15:04:05"),
			IsOwn:      msg.SenderID == currentUserIDStr,
		})
	}

	return result, nil
}
