package domain

import (
	"context"
	"time"
)

type Subject struct {
	SubjectCode string    `gorm:"primaryKey;type:varchar(5);not null;" json:"subject_code" valid:"required~Subject code is required"`
	Name        string    `gorm:"type:varchar(100);not null;" json:"name" valid:"required~Subject name is required"`
	Grade       int       `gorm:"not null" json:"grade" valid:"required~Grade is required"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type Student struct {
	StudentNSN string    `gorm:"primaryKey;type:varchar(10);not null;" json:"student_nsn" valid:"required~NSN is required"`
	Name       string    `gorm:"type:varchar(150);not null;" json:"name" valid:"required~Name is required"`
	Grade      int       `gorm:"not null" json:"grade" valid:"required~Grade is required"`
	GradeLabel string    `gorm:"type:varchar(5);not null;" json:"grade_label"`
	Gender     string    `gorm:"type:gender_enum;not null" json:"gender" valid:"required~Gender is required,in(male|female)~Invalid gender"`
	Telephone  string    `gorm:"type:varchar(13);not null;" json:"telephone" valid:"required~Telephone is required"`
	ParentID   int       `json:"parent_id"`
	Parent     Parent    `gorm:"references:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"parent" valid:"-"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type TestScore struct {
	TestScoreID int        `gorm:"primaryKey;autoIncrement" json:"test_score_id"`
	StudentNSN  string     `gorm:"not null" json:"student_nsn"`
	Student     Student    `gorm:"foreignKey:StudentNSN;references:StudentNSN;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"student"`
	SubjectCode string     `gorm:"not null" json:"subject_code"`
	Subject     Subject    `gorm:"foreignKey:SubjectCode;references:SubjectCode;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"subject"`
	UserID      int        `gorm:"not null" json:"user_id"`
	User        User       `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user"`
	Score       *float64   `json:"score" valid:"required~Score is required"`
	Type        *string    `gorm:"type:varchar(50);" json:"type" valid:"required~Type is required"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	SentAt      *time.Time `gorm:"index" json:"sent_at"`
}

type StudentRepo interface {
	GetAllStudent(ctx context.Context, userID int) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
	GetStudentByParentTelephone(ctx context.Context, parTel string) (*StudentsAssociateWithParent, error)
}

type StudentUseCase interface {
	GetAllStudent(ctx context.Context, userID int) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
	GetStudentByParentTelephone(ctx context.Context, parTel string) (*StudentsAssociateWithParent, error)
}
