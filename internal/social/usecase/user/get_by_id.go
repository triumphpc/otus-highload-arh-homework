package user

import (
	"context"
	"errors"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserUseCase struct {
	repo repository.UserRepository
}

func New(repo repository.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) GetByID(ctx context.Context, id int) (*entity.User, error) {
	if id < 1 {
		return nil, errors.New("empty user id")
	}

	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (uc *UserUseCase) Update(ctx context.Context, id string, updateData *entity.User) (*entity.User, error) {
	if id == "" {
		return nil, errors.New("empty user id")
	}

	// Валидация данных
	if updateData.FirstName == "" {
		return nil, errors.New("first name is required")
	}
	if updateData.LastName == "" {
		return nil, errors.New("last name is required")
	}

	// Получаем текущие данные пользователя
	existingUser, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Обновляем только разрешенные поля
	existingUser.FirstName = updateData.FirstName
	existingUser.LastName = updateData.LastName
	existingUser.BirthDate = updateData.BirthDate
	existingUser.Gender = updateData.Gender
	existingUser.Interests = updateData.Interests
	existingUser.City = updateData.City

	updatedUser, err := uc.repo.Update(ctx, id, existingUser)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (uc *UserUseCase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("empty user id")
	}

	// Проверя
