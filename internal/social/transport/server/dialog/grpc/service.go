package grpc

import (
	"context"
	"errors"
	"strconv"

	"otus-highload-arh-homework/internal/social/usecase/user"
	"otus-highload-arh-homework/pkg/proto/dialog/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DialogService struct {
	dialogv1.UnimplementedDialogServiceServer
	uc user.UserUseCase
}

func NewDialogService(uc user.UserUseCase) *DialogService {
	return &DialogService{uc: uc}
}

func (s *DialogService) SendMessage(ctx context.Context, req *dialogv1.SendMessageRequest) (*dialogv1.SendMessageResponse, error) {
	// Валидация
	if req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "message text cannot be empty")
	}

	senderID, err := strconv.Atoi(req.SenderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid sender ID")
	}

	receiverID, err := strconv.Atoi(req.ReceiverId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid receiver ID")
	}

	// Вызов use case
	err = s.uc.SendDialogMessage(ctx, int64(senderID), int64(receiverID), req.Text)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to send message")
	}

	return &dialogv1.SendMessageResponse{
		Success: true,
	}, nil
}

func (s *DialogService) GetMessages(ctx context.Context, req *dialogv1.GetMessagesRequest) (*dialogv1.GetMessagesResponse, error) {
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID")
	}

	otherUserID, err := strconv.Atoi(req.OtherUserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid other user ID")
	}

	// Вызов use case
	messages, err := s.uc.GetDialogMessages(ctx, int64(userID), int64(otherUserID))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to get messages")
	}

	// Конвертация в protobuf
	pbMessages := make([]*dialogv1.DialogMessage, 0, len(messages))
	for _, msg := range messages {
		pbMessages = append(pbMessages, &dialogv1.DialogMessage{
			MessageId:  msg.ID,
			SenderId:   msg.SenderID,
			ReceiverId: msg.ReceiverID,
			Text:       msg.Text,
			SentAt:     timestamppb.New(msg.SentAt),
		})
	}

	return &dialogv1.GetMessagesResponse{
		Messages: pbMessages,
	}, nil
}
