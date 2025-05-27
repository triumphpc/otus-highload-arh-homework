package user

import (
	"context"
	"errors"

	"otus-highload-arh-homework/internal/social/entity"
	"otus-highload-arh-homework/internal/social/repository"
)

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
