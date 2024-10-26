package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
)

type LoginRequest struct {
	Username string `json:"username" valid:"required~Username is required"`
	Password string `json:"password" valid:"required~Password is required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	// Teaching []int  `json:"teaching"`
	jwt.RegisteredClaims
}

type AuthRepo interface {
	Login(ctx context.Context, data *LoginRequest) (*[]string, error)
}

type AuthUseCase interface {
	Login(ctx context.Context, data *LoginRequest) (*[]string, error)
}
