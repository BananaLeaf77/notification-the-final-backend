package domain

import (
	"context"
	"time"
)

type Student struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Class           string     `json:"class"`
	Gender          string     `json:"gender"`
	TelephoneNumber int64      `json:"telephone_number"`
	ParentID        int        `json:"parent_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}

type StudentRepo interface {
	CreateStudent(ctx context.Context, student *Student) error
	GetAllStudent() (*[]Student, error)
	GetStudentByID(id int) (*Student, error)
	UpdateStudent(newDataStudent *Student) error
	DeleteStudent(id int) error
}

type StudentUseCase interface {
	CreateStudentUC(student *Student) error
	GetAllStudentUC() (*[]Student, error)
	GetStudentByIDUC(id int) (*Student, error)
	UpdateStudentUC(newDataStudent *Student) error
	DeleteStudentUC(id int) error
}
