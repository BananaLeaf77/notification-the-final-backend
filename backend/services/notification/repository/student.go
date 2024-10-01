package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(database *pgxpool.Pool) domain.StudentRepo {
	return &studentRepository{
		db: database,
	}
}

func (sp *studentRepository) GetAllStudent(ctx context.Context) (*[]domain.Student, error) {
	query := `
		SELECT id, name, class, gender, telephone, parent_id, created_at, updated_at, deleted_at
		FROM students
		WHERE deleted_at IS NULL;
	`

	rows, err := sp.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not get all students: %v", err)
	}
	defer rows.Close()

	var students []domain.Student
	for rows.Next() {
		var student domain.Student
		var studentTelephoneINT string

		err := rows.Scan(&student.ID, &student.Name, &student.Class, &student.Gender, &studentTelephoneINT, &student.ParentID, &student.CreatedAt, &student.UpdatedAt, &student.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan student: %v", err)
		}

		v, err := strconv.Atoi(studentTelephoneINT)
		if err != nil {
			return nil, err
		}

		student.Telephone = v

		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return &students, nil
}

func (sp *studentRepository) DownloadInputDataTemplate(ctx context.Context) (*string, error) {
	filePath := "./template/input_data_template.csv"

	// Check if the file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
	}
	return &filePath, nil
}
