package domain

import "context"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
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
