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

func (mUC *senderUC) SendMass(ctx context.Context, nsnList *[]string, userID *int, subjectCode string) error {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()

	err := mUC.emailSMTPRepo.SendMass(ctx, nsnList, userID, subjectCode)
	if err != nil {
		return err
	}
	return nil
}

func (mUC *senderUC) SendTestScores(ctx context.Context, examType string) error {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()

	err := mUC.emailSMTPRepo.SendTestScores(ctx, examType)
	if err != nil {
		return err
	}
	return nil
}
