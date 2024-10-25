package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Subject struct {
	SubjectID int            `gorm:"primaryKey;autoIncrement" json:"subject_id"`
	Name      string         `gorm:"type:varchar(100);not null;unique" json:"name" valid:"required~Subject name is required"`
	Grade     int            `gorm:"not null" json:"grade" valid:"required~Grade is required"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Student struct {
	StudentID int            `gorm:"primaryKey;autoIncrement" json:"student_id"`
	Name      string         `gorm:"type:varchar(150);not null;unique" json:"name" valid:"required~Name is required"`
	Class     string         `gorm:"type:varchar(3);not null" json:"class"`
	Gender    string         `gorm:"type:gender_enum;not null" json:"gender" valid:"required~Gender is required,in(male|female)~Invalid gender"`
	Telephone string         `gorm:"type:varchar(15);not null;unique" json:"telephone" valid:"required~Telephone is required"`
	ParentID  int            `json:"parent_id"`
	Parent    Parent         `gorm:"references:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"parent" valid:"-"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type TestScore struct {
	TestScoreID int            `gorm:"primaryKey;autoIncrement" json:"test_score_id"`
	StudentID   int            `json:"student_id"`
	Student     Student        `gorm:"references:StudentID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"student"`
	SubjectID   int            `json:"subject_id"`
	Subject     Subject        `gorm:"references:SubjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"subject"`
	TeacherID   int            `json:"teacher_id"`
	Teacher     User           `gorm:"references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"teacher"`
	Score       *float64       `json:"score" valid:"required~Score is required"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type StudentRepo interface {
	GetAllStudent(ctx context.Context) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
}

type StudentUseCase interface {
	GetAllStudentUC(ctx context.Context) (*[]Student, error)
	DownloadInputDataTemplate(ctx context.Context) (*string, error)
}
