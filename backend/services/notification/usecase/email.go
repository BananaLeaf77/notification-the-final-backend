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

func (mUC *emailSMTPUC) SendMass(ctx context.Context, payload *[]domain.EmailSMTPData) error {
	ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	defer cancel()

	err := mUC.emailSMTPRepo.SendMass(ctx, payload)
	if err != nil {
		return err
	}
	return nil
}
