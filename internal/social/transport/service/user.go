package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/transport/clients/dialog/grpc"
	"otus-highload-arh-homework/internal/social/transport/dto"
	userUC "otus-highload-arh-homework/internal/social/usecase/user"
)

type userUserCase interface {
	GetByID(ctx context.Context, id int) (*entity.User, error)
	Search(ctx context.Context, firstName, lastName string) ([]*entity.User, error)
	SendDialogMessage(ctx context.Context, senderID, receiverID int64, text string) error
	GetDialogMessages(ctx context.Context, user1ID, user2ID int64) ([]*entity.DialogMessage, error)
}

type friendUseCase interface {
	AddFriend(ctx context.Context, userID, friendID int) error
	RemoveFriend(ctx context.Context, userID, friendID int) error
	CheckFriendship(ctx context.Context, userID, friendID int) (bool, error)
	GetFriendsIDs(ctx context.Context, userID int) ([]int, error)
}

type UserService struct {
	userUC       userUserCase
	friendUC     friendUseCase
	dialogClient *grpc.Client
}

func NewUserService(
	userUC userUserCase,
	friendUC friendUseCase,
	dialogClient *grpc.Client,
) *UserService {
	return &UserService{
		userUC:       userUC,
		friendUC:     friendUC,
		dialogClient: dialogClient,
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

func (s *UserService) SendDialogMessageV2(ctx context.Context, senderID int, receiverIDStr, text string) error {
	// Валидация параметров
	if strings.TrimSpace(text) == "" {
		return errors.New("message text cannot be empty")
	}

	if _, err := strconv.Atoi(receiverIDStr); err != nil {
		return fmt.Errorf("invalid receiver ID: %w", err)
	}

	// Вызов gRPC клиента
	err := s.dialogClient.SendMessage(ctx, strconv.Itoa(senderID), receiverIDStr, text)
	if err != nil {
		return fmt.Errorf("gRPC SendMessage failed: %w", err)
	}

	return nil
}

func (s *UserService) GetDialogMessagesV2(ctx context.Context, currentUserID int, otherUserIDStr string) ([]dto.DialogMessageV2, error) {
	// Валидация ID собеседника
	if _, err := strconv.Atoi(otherUserIDStr); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Вызов gRPC клиента
	messages, err := s.dialogClient.GetMessages(ctx, strconv.Itoa(currentUserID), otherUserIDStr)
	if err != nil {
		return nil, fmt.Errorf("gRPC GetMessages failed: %w", err)
	}

	// Конвертация protobuf -> DTO
	result := make([]dto.DialogMessageV2, 0, len(messages))
	for _, msg := range messages {
		result = append(result, dto.DialogMessageV2{
			ID:         msg.MessageId,
			SenderID:   msg.SenderId,
			ReceiverID: msg.ReceiverId,
			Text:       msg.Text,
			SentAt:     msg.SentAt.AsTime(),
			IsOwn:      msg.SenderId == strconv.Itoa(currentUserID),
		})
	}

	return result, nil
}
