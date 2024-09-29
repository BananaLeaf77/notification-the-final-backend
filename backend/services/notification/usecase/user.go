package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type userUC struct {
	userRepo domain.UserRepo
	TimeOut  time.Duration
}

func NewUserUseCase(repo domain.UserRepo, timeOut time.Duration) domain.UserRepo {
	return &userUC{
		userRepo: repo,
		TimeOut:  timeOut,
	}
}

func (u *userUC) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()
	v, err := u.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return v, nil
}
