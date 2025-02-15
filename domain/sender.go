package domain

import "context"

type SenderRepo interface {
	SendMass(ctx context.Context, nsnList *[]string, userID *int, subjectCode string) error
	SendTestScores(ctx context.Context, examType string) error
}

type SenderUseCase interface {
	SendMass(ctx context.Context, nsnList *[]string, userID *int, subjectCode string) error
	SendTestScores(ctx context.Context, examType string) error
}
