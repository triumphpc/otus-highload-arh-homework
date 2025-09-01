package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"otus-highload-arh-homework/internal/social/entity"
)

func (uc *UserUseCase) SendDialogMessage(ctx context.Context, senderID, receiverID int64, text string) error {
	_, err := uc.repo.StoreDialogMessage(ctx, senderID, receiverID, text)
	if err != nil {
		return err
	}

	// Ставим асинхронную таску в сервис счетчиков для пересчета
	err = uc.counterQueue.PushCounterRecalc(ctx, receiverID)

	return err
}

func (uc *UserUseCase) GetDialogMessages(ctx context.Context, user1ID, user2ID int64) ([]*entity.DialogMessage, error) {
	// todo вообще storage должен возвращать DAO
	// а тут уже все конвертации
	_, err := uc.repo.GetDialogMessages(ctx, user1ID, user2ID)

	return nil, err

}

func (uc *UserUseCase) UpdateDialogMessagesUnreadCounter(ctx context.Context, currentUserID int64) error {
	count, err := uc.repo.GetUnreadMessagesCount(ctx, currentUserID)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("unread_count_%d", currentUserID)
	err = uc.cacher.Set(ctx, key, count, time.Hour*24*30)

	log.Println("Set counter for", currentUserID, count)

	if err != nil {
		return errors.Join(fmt.Errorf("can't set unreaded error"), err)
	}

	return nil
}

func (uc *UserUseCase) DialogMessagesUnreadCount(ctx context.Context, currentUserID int64) (int64, error) {
	var count int
	key := fmt.Sprintf("unread_count_%d", currentUserID)
	err := uc.cacher.Get(ctx, key, &count)

	if err != nil {
		return 0, err
	}

	return int64(count), nil
}
