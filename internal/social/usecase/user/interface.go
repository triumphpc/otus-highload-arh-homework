package user

import (
	"context"

	"otus-highload-arh-homework/internal/social/entity"
)

type UseCase interface {
	Register(ctx context.Context, user *entity.User) error
	Login(ctx context.Context, email, password string) (*entity.User, error)
	GetByID(ctx context.Context, id int) (*entity.User, error)
}
