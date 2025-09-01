package user

import (
	"context"

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
