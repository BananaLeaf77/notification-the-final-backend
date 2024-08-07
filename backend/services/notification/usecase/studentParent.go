package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type studentParentUseCase struct {
	repo    domain.StudentParentRepo
	TimeOut time.Duration
}

func NewStudentParentUseCase(repo domain.StudentParentRepo, to time.Duration) domain.StudentParentUseCase {
	return &studentParentUseCase{
		repo:    repo,
		TimeOut: to,
	}
}

func (spu *studentParentUseCase) CreateStudentAndParentUC(ctx context.Context, req *domain.CreateStudentParentRequest) error {
	ctx, cancel := context.WithTimeout(ctx, spu.TimeOut)
	defer cancel()

	err := spu.repo.CreateStudentAndParent(ctx, req)
	if err != nil {
		return err
	}
	return nil
}
