package domain

import (
	"context"
	"time"
)

type User struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"password"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type UserRepo interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error
}

type UserUseCase interface {
	GetAllStaff(ctx context.Context) (*[]SafeStaffData, error)
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateStaff(ctx context.Context, id int, payload *User) error
	CreateStaff(ctx context.Context, payload *User) (*User, error)
	DeleteStaff(ctx context.Context, id int) error
}
