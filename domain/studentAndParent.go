package domain

import (
	"context"
	"time"
)

type StudentAndParent struct {
	Student Student `json:"student"`
	Parent  Parent  `json:"parent"`
}

type ParentDataChangeRequest struct {
	RequestID          int        `gorm:"primaryKey;autoIncrement" json:"request_id"`
	UserID             int        `json:"user_id"`
	User               User       `gorm:"foreignKey:UserID;references:UserID" json:"user"`
	OldParentTelephone string     `json:"old_parent_telephone,omitempty"`
	NewParentName      *string    `json:"new_parent_name,omitempty"`
	NewParentTelephone *string    `json:"new_parent_telephone,omitempty"`
	NewParentEmail     *string    `json:"new_parent_email,omitempty"`
	NewParentGender    *string    `gorm:"type:gender_enum" json:"new_parent_gender" valid:"required~Gender is required,in(male|female)~Invalid gender"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	IsReviewed         bool       `gorm:"default:false" json:"is_reviewed"`
	DeletedAt          *time.Time `gorm:"index" json:"deleted_at"`
}

type StudentParentRepo interface {
	GetStudentDetailsByID(ctx context.Context, nsn string) (*StudentAndParent, error)
	CreateStudentAndParent(ctx context.Context, req *StudentAndParent) (*string, *[]string)
	// DeleteStudentAndParent(ctx context.Context, id int) error
	// SPMassDelete(ctx context.Context, studentIDS *[]int) error
	UpdateStudentAndParent(ctx context.Context, nsn string, payload *StudentAndParent) (*string, *[]string)
	// GetClassIDByName(className string) (*int, error)

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	GetAllDataChangeRequestByID(ctx context.Context, dcrID int) (*ParentDataChangeRequest, error)
	GetAllDataChangeRequest(ctx context.Context) (*[]ParentDataChangeRequest, error)
	DataChangeRequest(ctx context.Context, datas ParentDataChangeRequest, userID int) error
	ApproveDCR(ctx context.Context, req map[string]interface{}) (*string, error)
	DeleteDCR(ctx context.Context, dcrID int) error
}

type StudentParentUseCase interface {
	GetStudentDetailsByID(ctx context.Context, nsn string) (*StudentAndParent, error)
	CreateStudentAndParentUC(ctx context.Context, req *StudentAndParent) (*string, *[]string)
	// DeleteStudentAndParent(ctx context.Context, id int) error
	// SPMassDelete(ctx context.Context, studentIDS *[]int) error
	UpdateStudentAndParent(ctx context.Context, nsn string, payload *StudentAndParent) (*string, *[]string)
	// GetClassIDByName(className string) (*int, error)

	ImportCSV(ctx context.Context, payload *[]StudentAndParent) (*[]string, error)
	GetAllDataChangeRequestByID(ctx context.Context, dcrID int) (*ParentDataChangeRequest, error)
	GetAllDataChangeRequest(ctx context.Context) (*[]ParentDataChangeRequest, error)
	DataChangeRequest(ctx context.Context, datas ParentDataChangeRequest, userID int) error
	ApproveDCR(ctx context.Context, req map[string]interface{}) (*string, error)
	DeleteDCR(ctx context.Context, dcrID int) error
}
