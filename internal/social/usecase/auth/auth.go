package auth

import (
	"context"
	"errors"

	"otus-highload-arh-homework/internal/social/entity"
	repository2 "otus-highload-arh-homework/internal/social/repository"
)

type passwordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}

type AuthUseCase struct {
	repo   repository2.UserRepository
	hasher passwordHasher
}

func NewAuth(repo repository2.UserRepository, hasher passwordHasher) *AuthUseCase {
	return &AuthUseCase{repo: repo, hasher: hasher}
}

func (uc *AuthUseCase) Register(ctx context.Context, user *entity.User, password string) (*entity.User, error) {
	// 1. Проверка бизнес-правил
	if !user.IsAdult() {
		return nil, entity.ErrUnderageUser
	}

	// 2. Проверка уникальности email
	if _, err := uc.repo.GetByEmail(ctx, user.Email); !errors.Is(err, repository2.ErrUserNotFound) {
		return nil, repository2.ErrUserAlreadyExists
	}

	// 3. Хеширование пароля
	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	// 4. Сохранение
	if err := uc.repo.Create(ctx, user, hash); err != nil {
		return nil, err
	}

	return user, nil
}
