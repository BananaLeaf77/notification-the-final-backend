package domain

import "context"

type StudentAndParent struct {
	Student   Student `gorm:"foreignKey:StudentID;references:StudentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Parent    Parent  `gorm:"foreignKey:ParentID;references:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type StudentParentRepo interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) *[]string
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, payload *StudentAndParent) error

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
}

type StudentParentUseCase interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) *[]string
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, payload *StudentAndParent) error

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
}
