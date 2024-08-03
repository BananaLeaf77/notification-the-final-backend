package usecase

import (
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

func (sUC *studentUC) CreateStudentUC(student *domain.Student) error {

}

func (sUC *studentUC) GetAllStudentUC() (*[]domain.Student, error) {

}

func (sUC *studentUC) GetStudentByIDUC(id int) (*domain.Student, error) {

}

func (sUC *studentUC) UpdateStudentUC(newDataStudent *domain.Student) error {

}

func (sUC *studentUC) DeleteStudentUC(id int) error {

}
