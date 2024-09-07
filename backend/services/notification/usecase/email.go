package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type emailSMTPUC struct {
	emailSMTPRepo domain.EmailSMTPUseCase
	TimeOut       time.Duration
}

func NewMailSMTPUseCase(repo domain.EmailSMTPRepo, timeOut time.Duration) domain.EmailSMTPUseCase {
	return &emailSMTPUC{
		emailSMTPRepo: repo,
		TimeOut:       timeOut,
	}
}

func (mUC *emailSMTPUC) SendMass(ctx context.Context, idList *[]int) error {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()

	err := mUC.emailSMTPRepo.SendMass(ctx, idList)
	if err != nil {
		return err
	}
	return nil
}
