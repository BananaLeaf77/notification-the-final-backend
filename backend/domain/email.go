package domain

import "context"

type SenderRepo interface {
	SendMass(ctx context.Context, idList *[]int) error
}

type SenderUseCase interface {
	SendMass(ctx context.Context, idList *[]int) error
}
