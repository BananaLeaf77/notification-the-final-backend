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

// GAE CRUD NE NASKLENG!!!

type UserRepo interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	CreateStaff(ctx context.Context, payload *User) (*User, error)
}

type UserUseCase interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
	CreateStaff(ctx context.Context, payload *User) (*User, error)
}
