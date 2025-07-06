package auth

import (
	"context"
	"errors"
	"fmt"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository"
)

type passwordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}

type emailCasher interface {
	HasEmail(ctx context.Context, email string) (bool, error)
	SetEmail(ctx context.Context, email string) error
}

type AuthUseCase struct {
	repo        repository.UserRepository
	hasher      passwordHasher
	emailCasher emailCasher
}

func NewAuth(repo repository.UserRepository, hasher passwordHasher, casher emailCasher) *AuthUseCase {
	return &AuthUseCase{repo: repo, hasher: hasher, emailCasher: casher}
}

func (uc *AuthUseCase) Register(ctx context.Context, user *entity.User, password string) (*entity.User, error) {
	// 1. Проверка бизнес-правил
	if !user.IsAdult() {
		return nil, entity.ErrUnderageUser
	}

	// 2. Проверим, есть ли уже такой в KV хранилище
	ok, err := uc.emailCasher.HasEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}

	if ok {
		return nil, repository.ErrUserAlreadyExists
	}

	// 2. Проверка уникальности email
	// переведено н KV проверку
	//if _, err := uc.repo.GetByEmail(ctx, user.Email); !errors.Is(err, repository.ErrUserNotFound) {
	//	return nil, repository.ErrUserAlreadyExists
	//}

	// 3. Хеширование пароля
	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	// 4. Сохранение
	if err := uc.repo.Create(ctx, user, hash); err != nil {
		return nil, err
	}

	err = uc.emailCasher.SetEmail(ctx, user.Email)
	if err != nil {
		return nil, fmt.Errorf("set email in cache: %w", err)
	}

	return user, nil
}

// Login выполняет аутентификацию пользователя
func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	// 1. Находим пользователя по email
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 2. Проверяем пароль
	if !uc.hasher.Check(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
