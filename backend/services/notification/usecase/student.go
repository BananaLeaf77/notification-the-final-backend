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

func (sUC *studentUC) CreateStudentUC(ctx context.Context, student *domain.Student) error {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	err := sUC.studentRepo.CreateStudent(ctx, student)
	if err != nil {
		return err
	}
	return nil
}

func (sUC *studentUC) GetAllStudentUC(ctx context.Context) (*[]domain.Student, error) {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	students, err := sUC.studentRepo.GetAllStudent(ctx)
	if err != nil {
		return nil, err
	}
	return students, nil
}

func (sUC *studentUC) GetStudentByIDUC(ctx context.Context, id int) (*domain.StudentAndParent, error) {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	student, err := sUC.studentRepo.GetStudentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return student, nil
}

func (sUC *studentUC) UpdateStudentUC(ctx context.Context, newDataStudent *domain.Student) error {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	err := sUC.studentRepo.UpdateStudent(ctx, newDataStudent)
	if err != nil {
		return err
	}
	return nil
}

func (sUC *studentUC) DeleteStudentUC(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, sUC.TimeOut)
	defer cancel()

	err := sUC.studentRepo.DeleteStudent(ctx, id)
	if err != nil {
		return err
	}
	return nil
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
