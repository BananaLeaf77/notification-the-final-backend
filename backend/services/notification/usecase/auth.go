package usecase

import (
	"context"
	"notification/domain"
)

type authUC struct {
	authRepo domain.AuthRepo
}

func NewAuthUseCase(repo domain.AuthRepo) domain.AuthRepo {
	return &authUC{
		authRepo: repo,
	}
}

func (auc *authUC) Login(ctx context.Context, data *domain.LoginRequest) (*[]string, error) {
	datas, err := auc.authRepo.Login(ctx, data)
	if err != nil {
		return nil, err
	}
	return datas, nil
}
