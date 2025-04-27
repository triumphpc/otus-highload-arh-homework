package user

import (
	"context"
	"errors"
	"time"

	"otus-highload-arh-homework/internal/social/entity"
	repository2 "otus-highload-arh-homework/internal/social/repository"
)

type useCase struct {
	repo repository2.UserRepository
	// Добавим crypto-сервис для хеширования/проверки паролей
	hasher PasswordHasher
}

// New создает экземпляр UseCase
func New(repo repository2.UserRepository, hasher PasswordHasher) UseCase {
	return &useCase{
		repo:   repo,
		hasher: hasher,
	}
}

func (uc *useCase) Register(ctx context.Context, user *entity.User) error {
	// Валидация
	if user.FirstName == "" || user.LastName == "" {
		return entity.ErrInvalidUserName
	}

	if !user.IsAdult() {
		return entity.ErrUnderageUser
	}

	// Проверка уникальности email
	if _, err := uc.repo.GetByEmail(ctx, user.Email); !errors.Is(err, repository2.ErrUserNotFound) {
		return repository2.ErrUserAlreadyExists
	}

	// Хеширование пароля
	hashedPassword, err := uc.hasher.Hash(user.Password)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword

	user.CreatedAt = time.Now()
	return uc.repo.Create(ctx, user)
}

func (uc *useCase) Login(ctx context.Context, email, password string) (*entity.User, error) {
	// Получаем пользователя
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err // ErrUserNotFound или другая ошибка
	}

	// Проверяем пароль
	if !uc.hasher.Check(password, user.PasswordHash) {
		return nil, entity.ErrInvalidCredentials
	}

	return user, nil
}

func (uc *useCase) GetByID(ctx context.Context, id int) (*entity.User, error) {
	if id <= 0 {
		return nil, entity.ErrInvalidID
	}
	return uc.repo.GetByID(ctx, id)
}
