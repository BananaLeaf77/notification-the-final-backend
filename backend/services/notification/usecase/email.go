package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type mailgunUC struct {
	mailGunRepo domain.MailGunRepo
	TimeOut     time.Duration
}

func NewMailGunUseCase(repo domain.MailGunRepo, timeOut time.Duration) domain.MailGunUseCase {
	return &mailgunUC{
		mailGunRepo: repo,
		TimeOut:     timeOut,
	}
}

func (mUC *mailgunUC) SendMass(ctx context.Context, studentName, parentName, emailAddress string) error {
	ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	defer cancel()

	err := mUC.mailGunRepo.SendMass(ctx, studentName, parentName, emailAddress)
	if err != nil {
		return err
	}
	return nil
}
