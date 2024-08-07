package domain

import "context"

type CreateStudentParentRequest struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type StudentParentRepo interface {
	CreateStudentAndParent(ctx context.Context, req *CreateStudentParentRequest) error
}

type StudentParentUseCase interface {
	CreateStudentAndParentUC(ctx context.Context, req *CreateStudentParentRequest) error
}
