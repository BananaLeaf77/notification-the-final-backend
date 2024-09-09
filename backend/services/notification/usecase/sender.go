package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type senderUC struct {
	emailSMTPRepo domain.SenderRepo
	TimeOut       time.Duration
}

func NewSenderUseCase(repo domain.SenderRepo, timeOut time.Duration) domain.SenderRepo {
	return &senderUC{
		emailSMTPRepo: repo,
		TimeOut:       timeOut,
	}
}

func (mUC *senderUC) SendMass(ctx context.Context, idList *[]int) error {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()

	err := mUC.emailSMTPRepo.SendMass(ctx, idList)
	if err != nil {
		return err
	}
	return nil
}
