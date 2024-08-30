package domain

import "context"

type EmailSMTPData struct {
	StudentName  string `json:"student_name"`
	ParentName   string `json:"parent_name"`
	EmailAddress string `json:"email"`
}

type EmailSMTPRepo interface {
	SendMass(ctx context.Context, payloadList *[]EmailSMTPData) error
}

type EmailSMTPUseCase interface {
	SendMass(ctx context.Context, payloadList *[]EmailSMTPData) error
}
