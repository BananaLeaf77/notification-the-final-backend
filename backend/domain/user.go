package domain

import "context"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserRepo interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
}

type UserUseCase interface {
	FindUserByUsername(ctx context.Context, username string) (*User, error)
}
