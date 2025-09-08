package user

import (
	"context"
	"errors"
	"time"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository"
)

type counterQueue interface {
	PushCounterRecalc(ctx context.Context, userID int64) error
}

type cacher interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
}

type UserUseCase struct {
	repo         repository.UserRepository
	counterQueue counterQueue
	cacher       cacher
}

func New(repo repository.UserRepository, counterQueue counterQueue, cacher cacher) *UserUseCase {
	return &UserUseCase{repo: repo, counterQueue: counterQueue, cacher: cacher}
}

func (uc *UserUseCase) GetByID(ctx context.Context, id int) (*entity.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// Search - поиск по имении фамилии
func (uc *UserUseCase) Search(ctx context.Context, firstName, lastName string) ([]*entity.User, error) {
	users, err := uc.repo.Search(ctx, firstName, lastName)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return users, nil
}
