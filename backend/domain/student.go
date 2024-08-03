package domain

import (
	"context"
	"time"
)

type Student struct {
	ID              int        `json:"id"`
	Name            string     `json:"name" valid:"required~Name is required"`
	Class           string     `json:"class" valid:"required~Class is required"`
	Gender          string     `json:"gender" valid:"required~Gender is required"`
	TelephoneNumber int64      `json:"telephone_number" valid:"required~Telephone Number is required"`
	ParentID        int        `json:"parent_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}

type StudentRepo interface {
	CreateStudent(ctx context.Context, student *Student) error
	GetAllStudent(ctx context.Context) (*[]Student, error)
	GetStudentByID(ctx context.Context, id int) (*Student, error)
	UpdateStudent(ctx context.Context, newDataStudent *Student) error
	DeleteStudent(ctx context.Context, id int) error
}

type StudentUseCase interface {
	CreateStudentUC(ctx context.Context, student *Student) error
	GetAllStudentUC(ctx context.Context) (*[]Student, error)
	GetStudentByIDUC(ctx context.Context, id int) (*Student, error)
	UpdateStudentUC(ctx context.Context, newDataStudent *Student) error
	DeleteStudentUC(ctx context.Context, id int) error
}
