package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	Username  string         `gorm:"unique;not null"`
	Password  string         `gorm:"not null"`
	Role      string         `gorm:"not null";type:role_enum;`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserRepo interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error
}

type UserUseCase interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	GetStaffDetail(ctx context.Context, id int) (*SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error
}
