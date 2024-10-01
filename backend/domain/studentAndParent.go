package domain

import "context"

type StudentAndParent struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type StudentParentRepo interface {
	GetStudentDetailByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) error
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) error

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
}

type StudentParentUseCase interface {
	GetStudentDetailByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) error
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) error

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
}
