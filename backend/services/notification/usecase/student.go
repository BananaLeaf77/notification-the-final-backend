package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type studentUC struct {
	studentRepo domain.StudentRepo
	TimeOut     time.Duration
}

func NewStudentUseCase(repo domain.StudentRepo, timeOut time.Duration) domain.StudentUseCase {
	return &studentUC{
		studentRepo: repo,
		TimeOut:     timeOut,
	}
}

func (sUC *studentUC) GetAllStudent(ctx context.Context, userID int) (*[]domain.Student, error) {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	students, err := sUC.studentRepo.GetAllStudent(ctx, userID)
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (sUC *studentUC) DownloadInputDataTemplate(ctx context.Context) (*string, error) {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	filepath, err := sUC.studentRepo.DownloadInputDataTemplate(ctx)
	if err != nil {
		return nil, err
	}
	return filepath, nil
}
