package repository

import (
	"database/sql"
	"fmt"
	"notification/domain"
	"time"
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
	// Check for duplicate student
	duplicateCheckQuery := `
		SELECT id FROM students
		WHERE name = $1 AND class = $2 AND gender = $3 AND telephone_number = $4;
	`
	var existingID int

	err := sp.db.QueryRow(duplicateCheckQuery, student.Name, student.Class, student.Gender, student.TelephoneNumber).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("could not check for duplicate student: %v", err)
	}

	if existingID != 0 {
		return fmt.Errorf("student already exists with ID: %d", existingID)
	}

	// Insert new student
	insertQuery := `
		INSERT INTO students (name, class, gender, telephone_number, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	now := time.Now()

	var id int
	err = sp.db.QueryRow(insertQuery, student.Name, student.Class, student.Gender, student.TelephoneNumber, student.ParentID, now, now).Scan(&id)
	if err != nil {
		return fmt.Errorf("could not insert student: %v", err)
	}

	student.ID = id
	student.CreatedAt = now
	student.UpdatedAt = now

	return nil
}

func (sp *studentRepository) GetAllStudent() (*[]domain.Student, error) {
	query := `
		SELECT id, name, class, gender, telephone_number, parent_id, created_at, updated_at, deleted_at
		FROM students
		WHERE deleted_at IS NULL;
	`

	rows, err := sp.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("could not get all students: %v", err)
	}
	defer rows.Close()

	var students []domain.Student
	for rows.Next() {
		var student domain.Student
		err := rows.Scan(&student.ID, &student.Name, &student.Class, &student.Gender, &student.TelephoneNumber, &student.ParentID, &student.CreatedAt, &student.UpdatedAt, &student.DeletedAt)
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

func (sp *studentRepository) GetStudentByID(id int) (*domain.Student, error) {
	query := `
		SELECT id, name, class, gender, telephone_number, parent_id, created_at, updated_at, deleted_at
		FROM students
		WHERE id = $1 AND deleted_at IS NULL;
	`

	var student domain.Student
	err := sp.db.QueryRow(query, id).Scan(&student.ID, &student.Name, &student.Class, &student.Gender, &student.TelephoneNumber, &student.ParentID, &student.CreatedAt, &student.UpdatedAt, &student.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("student not found")
		}
		return nil, fmt.Errorf("could not get student: %v", err)
	}

	return &student, nil
}

func (sp *studentRepository) UpdateStudent(student *domain.Student) error {
	query := `
		UPDATE students
		SET name = $1, class = $2, gender = $3, telephone_number = $4, parent_id = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL;
	`

	now := time.Now()
	_, err := sp.db.Exec(query, student.Name, student.Class, student.Gender, student.TelephoneNumber, student.ParentID, now, student.ID)
	if err != nil {
		return fmt.Errorf("could not update student: %v", err)
	}

	student.UpdatedAt = now
	return nil
}

func (sp *studentRepository) DeleteStudent(id int) error {
	query := `
		UPDATE students
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL;
	`

	now := time.Now()
	_, err := sp.db.Exec(query, now, id)
	if err != nil {
		return fmt.Errorf("could not delete student: %v", err)
	}

	return nil
}
