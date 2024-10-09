package domain

import "context"

type StudentAndParent struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type StudentParentRepo interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) error
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, payload *StudentAndParent) error

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
}

type StudentParentUseCase interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) error
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, payload *StudentAndParent) error

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
}
