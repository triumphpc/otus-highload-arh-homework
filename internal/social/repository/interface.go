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
	AddFriend(ctx context.Context, userID, friendID int) error
	RemoveFriend(ctx context.Context, userID, friendID int) error
	CheckFriendship(ctx context.Context, userID, friendID int) (bool, error)
}

// PostRepository определяет контракт для работы с хранилищем постов
type PostRepository interface {
	Create(ctx context.Context, post *entity.Post) (string, error)
	Update(ctx context.Context, post *entity.Post) error
	Delete(ctx context.Context, postID string) error
	Get(ctx context.Context, postID string) (*entity.Post, error)
	GetFeed(ctx context.Context, userID, offset, limit int) ([]*entity.Post, error)
}
