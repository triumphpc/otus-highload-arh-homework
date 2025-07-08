package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository"
)

type passwordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}

type emailCasher interface {
	HasEmail(ctx context.Context, email string) (bool, error)
	DeleteEmail(ctx context.Context, email string) error
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

	go func(user entity.User, password string) {
		var err error
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer func() {
			if err != nil {
				delErr := uc.emailCasher.DeleteEmail(ctx, user.Email)
				err = errors.Join(err, delErr)
				log.Println(fmt.Errorf("email check failed: %w", err))
			}

			cancel()
		}()
		// 3. Асинхронное сохранение в основном storage
		hash, err := uc.hasher.Hash(password)
		if err != nil {
			return
		}

		// 4. Сохранение
		if err := uc.repo.Create(ctx, &user, hash); err != nil {
			return
		}

		log.Println("user registered async")
	}(*user, password)

	log.Println("user registered done")

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
