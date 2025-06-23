package repository

import (
	"context"

	"otus-highload-arh-homework/internal/social/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User, passwordHash string) error
	GetByID(ctx context.Context, id int) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Search(ctx context.Context, firstName, lastName string) ([]*entity.User, error)
	StoreDialogMessage(ctx context.Context, senderID, recipientID int64, content string) (int64, error)
	GetDialogMessages(ctx context.Context, senderID, recipientID int64) ([]*entity.DialogMessage, error)
}
