package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID    int            `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username  string         `gorm:"type:varchar(100);not null;" json:"username"`
	Name      string         `gorm:"type:varchar(100);not null;" json:"name"`
	Password  string         `gorm:"type:varchar(100);not null" json:"password"`
	Role      string         `gorm:"type:varchar(10);not null" json:"role"`
	Teaching  []*Subject     `gorm:"many2many:user_subjects" json:"teaching"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Profile struct {
	UserID   int    `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username string `gorm:"type:varchar(100);not null;" json:"username"`
	Password string `gorm:"type:varchar(100);not null" json:"password"`
	Role     string `gorm:"type:varchar(10);not null" json:"role"`
}

type UserRepo interface {
	GetAdminByAdmin(ctx context.Context) (*SafeStaffData, error)
	ShowProfile(ctx context.Context, uID int) (*SafeStaffData, error)
	// Staff
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User, subjectIDs []int) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error
	DeleteStaffMass(ctx context.Context, ids *[]int) error

	// Subject
	CreateSubject(ctx context.Context, subject *Subject) error
	CreateSubjectBulk(ctx context.Context, subjects *[]Subject) (*[]string, error)
	GetAllSubject(ctx context.Context, userID int) (*[]Subject, error)
	UpdateSubject(ctx context.Context, id int, newSubjectData *Subject) error
	DeleteSubject(ctx context.Context, id int) error
	DeleteSubjectMass(ctx context.Context, ids *[]int) error
	GetSubjectsForTeacher(ctx context.Context, userID int) (*SafeStaffData, error)
	GetSubjectDetail(ctx context.Context, subjectID int) (*Subject, error)

	// TestScore
	InputTestScores(ctx context.Context, teacherID int, testScores *InputTestScorePayload) error
	GetAllTestScores(ctx context.Context) (*[]TestScore, error)
	GetAllTestScoresBySubjectID(ctx context.Context, subjectID int) (*[]TestScore, error)
}

type UserUseCase interface {
	GetAdminByAdmin(ctx context.Context) (*SafeStaffData, error)
	ShowProfile(ctx context.Context, uID int) (*SafeStaffData, error)

	// Staff
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User, subjectIDs []int) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error
	DeleteStaffMass(ctx context.Context, ids *[]int) error

	// Subject
	CreateSubject(ctx context.Context, subject *Subject) error
	CreateSubjectBulk(ctx context.Context, subjects *[]Subject) (*[]string, error)
	GetAllSubject(ctx context.Context, userID int) (*[]Subject, error)
	UpdateSubject(ctx context.Context, id int, newSubjectData *Subject) error
	DeleteSubject(ctx context.Context, id int) error
	DeleteSubjectMass(ctx context.Context, ids *[]int) error
	GetSubjectsForTeacher(ctx context.Context, userID int) (*SafeStaffData, error)
	GetSubjectDetail(ctx context.Context, subjectID int) (*Subject, error)

	// TestScore
	InputTestScores(ctx context.Context, teacherID int, testScores *InputTestScorePayload) error
	GetAllTestScores(ctx context.Context) (*[]TestScore, error)
	GetAllTestScoresBySubjectID(ctx context.Context, subjectID int) (*[]TestScore, error)
}
