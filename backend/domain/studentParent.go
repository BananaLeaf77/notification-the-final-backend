package domain

import "context"

type StudentAndParent struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type StudentParentRepo interface {
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) error
	GetStudentAndParent(ctx context.Context, studentID string) (*StudentAndParent, error)
	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) error
}

type StudentParentUseCase interface {
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) error
	GetStudentAndParent(ctx context.Context, studentID string) (*StudentAndParent, error)
	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) error
}
