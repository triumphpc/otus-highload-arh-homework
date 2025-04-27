package repository

import (
	"context"

	"otus-highload-arh-homework/internal/social/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User, passwordHash string) error
	GetByID(ctx context.Context, id int) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}
