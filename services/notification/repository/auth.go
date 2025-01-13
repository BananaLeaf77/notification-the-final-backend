package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"notification/middleware"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) domain.AuthRepo {
	return &authRepository{
		db: db,
	}
}

func (ar *authRepository) Login(ctx context.Context, data *domain.LoginRequest) (*[]string, error) {
	var user domain.User
	var dataList []string

	err := ar.db.Where("username = ?", data.Username).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	dataList = append(dataList, user.Role, user.Username)

	token, err := middleware.GenerateJWT(user.UserID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token, err : %v", err)
	}

	dataList = append(dataList, token)

	return &dataList, nil
}
