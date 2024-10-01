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

func (u *userUC) CreateStaff(ctx context.Context, payload *domain.User) (*domain.User, error) {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()
	v, err := u.userRepo.CreateStaff(ctx, payload)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) GetAllStaff(ctx context.Context) (*[]domain.SafeStaffData, error) {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()
	v, err := u.userRepo.GetAllStaff(ctx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) DeleteStaff(ctx context.Context, id int) error {
	err := u.userRepo.DeleteStaff(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) UpdateStaff(ctx context.Context, id int, payload *domain.User) error {
	err := u.userRepo.UpdateStaff(ctx, id, payload)
	if err != nil {
		return err
	}

	return nil
}

