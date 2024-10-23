package domain

import (
	"context"
	"time"
)

type StudentAndParent struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type DataChangeRequest struct {
	RequestID           int       `gorm:"primaryKey;autoIncrement" json:"request_id"`
	OldStudentName      *string   `json:"old_student_name,omitempty"`
	OldStudentTelephone *string   `json:"old_student_telephone,omitempty"`
	OldParentName       *string   `json:"old_parent_name,omitempty"`
	OldParentTelephone  *string   `json:"old_parent_telephone,omitempty"`
	OldParentEmail      *string   `json:"old_parent_email,omitempty"`
	NewStudentName      *string   `json:"new_student_name,omitempty"`
	NewStudentTelephone *string   `json:"new_student_telephone,omitempty"`
	NewParentName       *string   `json:"new_parent_name,omitempty"`
	NewParentTelephone  *string   `json:"new_parent_telephone,omitempty"`
	NewParentEmail      *string   `json:"new_parent_email,omitempty"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	IsReviewed          bool      `gorm:"default:false" json:"is_reviewed"`
}

type StudentParentRepo interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) *[]string
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) error
	GetClassIDByName(className string) (*int, error)

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	DataChangeRequest(ctx context.Context, datas DataChangeRequest) error
}

type StudentParentUseCase interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) *[]string
	DeleteStudentAndParent(ctx context.Context, id int) error
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) error
	GetClassIDByName(className string) (*int, error)

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	DataChangeRequest(ctx context.Context, datas DataChangeRequest) error
}
