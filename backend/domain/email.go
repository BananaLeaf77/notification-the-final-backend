package domain

import "context"

type EmailSMTPData struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type EmailSMTPRepo interface {
	SendMass(ctx context.Context, idList *[]int) error
}

type EmailSMTPUseCase interface {
	SendMass(ctx context.Context, idList *[]int) error
}
