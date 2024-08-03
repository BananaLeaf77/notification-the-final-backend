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

}

func (sUC *studentUC) GetAllStudentUC(ctx context.Context) (*[]domain.Student, error) {

}

func (sUC *studentUC) GetStudentByIDUC(ctx context.Context, id int) (*domain.Student, error) {

}

func (sUC *studentUC) UpdateStudentUC(ctx context.Context, newDataStudent *domain.Student) error {

}

func (sUC *studentUC) DeleteStudentUC(ctx context.Context, id int) error {

}
