package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"time"

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

func (sp *studentRepository) CreateStudent(ctx context.Context, student *domain.Student) error {
	duplicateCheckQuery := `
		SELECT id FROM students
		WHERE name = $1 AND class = $2 AND gender = $3 AND telephone_number = $4;
	`
	var existingID int

	err := sp.db.QueryRow(ctx, duplicateCheckQuery, student.Name, student.Class, student.Gender, student.Telephone).Scan(&existingID)
	if err != nil && err.Error() != "no rows in result set" {
		return fmt.Errorf("could not check for duplicate student: %v", err)
	}

	if existingID != 0 {
		return fmt.Errorf("student already exists with ID: %d", existingID)
	}

	insertQuery := `
		INSERT INTO students (name, class, gender, telephone_number, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	now := time.Now()

	var id int
	err = sp.db.QueryRow(ctx, insertQuery, student.Name, student.Class, student.Gender, student.Telephone, student.ParentID, now, now).Scan(&id)
	if err != nil {
		return fmt.Errorf("could not insert student: %v", err)
	}

	student.ID = id
	student.CreatedAt = now
	student.UpdatedAt = now

	return nil
}

func (sp *studentRepository) GetAllStudent(ctx context.Context) (*[]domain.Student, error) {
	query := `
		SELECT id, name, class, gender, telephone_number, parent_id, created_at, updated_at, deleted_at
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
		err := rows.Scan(&student.ID, &student.Name, &student.Class, &student.Gender, &student.Telephone, &student.ParentID, &student.CreatedAt, &student.UpdatedAt, &student.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan student: %v", err)
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return &students, nil
}

func (sp *studentRepository) GetStudentByID(ctx context.Context, id int) (*domain.Student, error) {
	query := `
		SELECT id, name, class, gender, telephone_number, parent_id, created_at, updated_at, deleted_at
		FROM students
		WHERE id = $1 AND deleted_at IS NULL;
	`

	var student domain.Student
	err := sp.db.QueryRow(ctx, query, id).Scan(&student.ID, &student.Name, &student.Class, &student.Gender, &student.Telephone, &student.ParentID, &student.CreatedAt, &student.UpdatedAt, &student.DeletedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("student not found")
		}
		return nil, fmt.Errorf("could not get student: %v", err)
	}

	return &student, nil
}

func (sp *studentRepository) UpdateStudent(ctx context.Context, student *domain.Student) error {
	query := `
		UPDATE students
		SET name = $1, class = $2, gender = $3, telephone_number = $4, parent_id = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL;
	`

	now := time.Now()
	_, err := sp.db.Exec(ctx, query, student.Name, student.Class, student.Gender, student.Telephone, student.ParentID, now, student.ID)
	if err != nil {
		return fmt.Errorf("could not update student: %v", err)
	}

	student.UpdatedAt = now
	return nil
}

func (sp *studentRepository) DeleteStudent(ctx context.Context, id int) error {
	query := `
		UPDATE students
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL;
	`

	now := time.Now()
	_, err := sp.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("could not delete student: %v", err)
	}

	return nil
}
