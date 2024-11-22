package domain

import "context"

type SenderRepo interface {
	SendMass(ctx context.Context, idList *[]int, userID *int, subjectID int) error
}

type SenderUseCase interface {
	SendMass(ctx context.Context, idList *[]int, userID *int, subjectID int) error
}
