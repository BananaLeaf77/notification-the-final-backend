package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID    int            `gorm:"primaryKey" json:"user_id"`
	Username  string         `gorm:"unique;not null" json:"username" valid:"required~Username is required"`
	Password  string         `gorm:"not null" json:"password" valid:"required~Password is required"`
	Role      string         `gorm:"not null;type:role_enum" json:"role"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type UserRepo interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error

	GetlAllClass(ctx context.Context) (*[]Class, error)
	CreateClass(ctx context.Context, classData *Class) error
	DeleteClass(ctx context.Context, id int) error
}

type UserUseCase interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error

	GetlAllClass(ctx context.Context) (*[]Class, error)
	CreateClass(ctx context.Context, classData *Class) error
	DeleteClass(ctx context.Context, id int) error
}
