package repository

import (
	"database/sql"
	"notification/domain"
)

type studentRepository struct {
	db *sql.DB
}

func NewStudentRepository(database *sql.DB) domain.StudentRepo {
	return &studentRepository{
		db: database,
	}
}

func (sp *studentRepository) CreateStudent(student *domain.Student) error {

}
