package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID    int            `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username  string         `gorm:"type:varchar(100);not null;unique" json:"username"`
	Password  string         `gorm:"type:varchar(100);not null" json:"password"`
	Role      string         `gorm:"type:varchar(10);not null" json:"role"`
	Teaching  []*Subject     `gorm:"many2many:user_subjects" json:"teaching"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type UserRepo interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User, subjectIDs []int) error
	CreateStaff(ctx context.Context, payload *User, subjectIDs []int) (*User, error)
	DeleteStaff(ctx context.Context, id int) error

	CreateSubject(ctx context.Context, subject *Subject) error
	CreateSubjectBulk(ctx context.Context, subjects *[]Subject) (*[]string, error)
	GetAllSubject(ctx context.Context) (*[]Subject, error)
	UpdateSubject(ctx context.Context, id int, newSubjectData *Subject) error
	DeleteSubject(ctx context.Context, id int) error

	GetSubjectsForTeacher(ctx context.Context, userID int) (*[]Subject, error)
	// InputTestScores(ctx context.Context)

	// GetlAllClass(ctx context.Context) (*[]Class, error)
	// CreateClass(ctx context.Context, classData *Class) error
	// DeleteClass(ctx context.Context, id int) error
}

type UserUseCase interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User, subjectIDs []int) error
	CreateStaff(ctx context.Context, payload *User, subjectIDs []int) (*User, error)
	DeleteStaff(ctx context.Context, id int) error

	CreateSubject(ctx context.Context, subject *Subject) error
	CreateSubjectBulk(ctx context.Context, subjects *[]Subject) (*[]string, error)
	GetAllSubject(ctx context.Context) (*[]Subject, error)
	UpdateSubject(ctx context.Context, id int, newSubjectData *Subject) error
	DeleteSubject(ctx context.Context, id int) error

	GetSubjectsForTeacher(ctx context.Context, userID int) (*[]Subject, error)

	// GetlAllClass(ctx context.Context) (*[]Class, error)
	// CreateClass(ctx context.Context, classData *Class) error
	// DeleteClass(ctx context.Context, id int) error
}
