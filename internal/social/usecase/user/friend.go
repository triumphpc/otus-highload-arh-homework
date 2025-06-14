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

// GetFriendsIDs возвращает только ID друзей пользователя
func (uc *FriendUseCase) GetFriendsIDs(ctx context.Context, userID int) ([]int, error) {
	// Проверяем существование пользователя
	if _, err := uc.userRepo.GetByID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// Получаем список ID друзей
	friendsIDs, err := uc.userRepo.GetFriendsIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends IDs: %w", err)
	}

	return friendsIDs, nil
}
