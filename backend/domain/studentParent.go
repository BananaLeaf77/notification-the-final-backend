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
	UpdateStudentandParent(ctx context.Context, id int64, student *Student, parent *Parent) error
}

type StudentParentUseCase interface {
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) error
	GetStudentAndParent(ctx context.Context, studentID string) (*StudentAndParent, error)
	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	UpdateStudentandParent(ctx context.Context, id int64, student *Student, parent *Parent) error
}
