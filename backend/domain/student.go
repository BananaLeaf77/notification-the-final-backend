package domain

import (
	"context"
	"time"
)

type Student struct {
	ID        int        `json:"id"`
	Name      string     `json:"name" valid:"required~Name is required"`
	Class     string     `json:"class" valid:"required~Class is required"`
	Gender    string     `json:"gender" valid:"required~Gender is required"`
	Telephone int        `json:"telephone" valid:"required~Telephone is required,numeric~Telephone must be a number"`
	ParentID  int        `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type StudentRepo interface {
	GetAllStudent(ctx context.Context) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
}

type StudentUseCase interface {
	GetAllStudentUC(ctx context.Context) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
}
