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
	RequestID          int       `gorm:"primaryKey;autoIncrement" json:"request_id"`
	OldParentTelephone string    `json:"old_parent_telephone,omitempty"`
	NewParentName      *string   `json:"new_parent_name,omitempty"`
	NewParentTelephone *string   `json:"new_parent_telephone,omitempty"`
	NewParentEmail     *string   `json:"new_parent_email,omitempty"`
	NewParentGender    *string   `gorm:"type:gender_enum" json:"new_parent_gender" valid:"required~Gender is required,in(male|female)~Invalid gender"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	IsReviewed         bool      `gorm:"default:false" json:"is_reviewed"`
}

type StudentParentRepo interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) *[]string
	DeleteStudentAndParent(ctx context.Context, id int) error
	SPMassDelete(ctx context.Context, studentIDS *[]int) error
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) *[]string
	// GetClassIDByName(className string) (*int, error)

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	GetAllDataChangeRequestByID(ctx context.Context, dcrID int) (*DataChangeRequest, error)
	GetAllDataChangeRequest(ctx context.Context) (*[]DataChangeRequest, error)
	DataChangeRequest(ctx context.Context, datas DataChangeRequest) error
	ReviewDCR(ctx context.Context, dcrID int) error
}

type StudentParentUseCase interface {
	GetStudentDetailsByID(ctx context.Context, id int) (*StudentAndParent, error)
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) *[]string
	DeleteStudentAndParent(ctx context.Context, id int) error
	SPMassDelete(ctx context.Context, studentIDS *[]int) error
	UpdateStudentAndParent(ctx context.Context, id int, payload *StudentAndParent) *[]string
	// GetClassIDByName(className string) (*int, error)

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	GetAllDataChangeRequest(ctx context.Context) (*[]DataChangeRequest, error)
	GetAllDataChangeRequestByID(ctx context.Context, dcrID int) (*DataChangeRequest, error)
	DataChangeRequest(ctx context.Context, datas DataChangeRequest) error
	ReviewDCR(ctx context.Context, dcrID int) error
}
