package user

import (
	"context"
	"errors"
	"fmt"

	"otus-highload-arh-homework/internal/social/repository"
)

type FriendUseCase struct {
	userRepo repository.UserRepository
}

func NewFriendUseCase(
	userRepo repository.UserRepository,
) *FriendUseCase {
	return &FriendUseCase{
		userRepo: userRepo,
	}
}

// AddFriend добавляет пользователя в друзья (бизнес-логика без транспорта)
func (uc *FriendUseCase) AddFriend(ctx context.Context, userID, friendID int) error {
	// Проверяем, что пользователь не пытается добавить самого себя
	if userID == friendID {
		return ErrSelfOperation
	}

	// Проверяем существование друга
	if _, err := uc.userRepo.GetByID(ctx, friendID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Проверяем, не являются ли уже друзьями
	areFriends, err := uc.userRepo.CheckFriendship(ctx, userID, friendID)
	if err != nil {
		return err
	}
	if areFriends {
		return ErrAlreadyFriends
	}

	// Добавляем в друзья
	return uc.userRepo.AddFriend(ctx, userID, friendID)
}

// RemoveFriend удаляет пользователя из друзей (бизнес-логика)
func (uc *FriendUseCase) RemoveFriend(ctx context.Context, userID, friendID int) error {
	if userID == friendID {
		return ErrSelfOperation
	}

	if _, err := uc.userRepo.GetByID(ctx, friendID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to check friend existence: %w", err)
	}

	areFriends, err := uc.userRepo.CheckFriendship(ctx, userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to check friendship: %w", err)
	}
	if !areFriends {
		return ErrNotFriends
	}

	if err := uc.userRepo.RemoveFriend(ctx, userID, friendID); err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	return nil
}

// CheckFriendship проверяет, дружат ли пользователи
func (uc *FriendUseCase) CheckFriendship(ctx context.Context, userID, friendID int) (bool, error) {
	return uc.userRepo.CheckFriendship(ctx, userID, friendID)
}
