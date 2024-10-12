package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Student struct {
	StudentID int            `gorm:"primaryKey;autoIncrement" json:"student_id"`
	Name      string         `gorm:"type:varchar(150);not null;unique" json:"name" valid:"required~Name is required"`
	Class     string         `gorm:"type:varchar(3);not null" json:"class" valid:"required~Class is required"`
	Gender    string         `gorm:"type:gender_enum;not null" json:"gender" valid:"required~Gender is required,in(male|female)~Invalid gender"`
	Telephone string         `gorm:"type:varchar(15);not null" json:"telephone" valid:"required~Telephone is required"`
	ParentID  int            `json:"parent_id"`
	Parent    Parent         `gorm:"references:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"parent"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type StudentRepo interface {
	GetAllStudent(ctx context.Context) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
}

type StudentUseCase interface {
	GetAllStudentUC(ctx context.Context) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
}
