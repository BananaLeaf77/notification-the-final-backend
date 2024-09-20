package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"os"
	"strconv"
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
		WHERE name = $1 AND class = $2 AND gender = $3 AND telephone = $4;
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
		INSERT INTO students (name, class, gender, telephone, parent_id, created_at, updated_at)
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

func (sp *studentRepository) GetStudentByID(ctx context.Context, id int) (*domain.StudentAndParent, error) {
	query := `
		SELECT s.id, s.name, s.class, s.gender, s.telephone, s.parent_id, s.created_at, s.updated_at, s.deleted_at,
		p.id, p.name, p.gender, p.telephone, p.email, p.created_at, p.updated_at, p.deleted_at
		FROM students s
		JOIN parents p ON s.parent_id = p.id
		WHERE s.id = $1 AND s.deleted_at IS NULL AND p.deleted_at IS NULL;
	`

	var result domain.StudentAndParent
	var pTelephone string
	var sTelephone string

	err := sp.db.QueryRow(ctx, query, id).Scan(
		&result.Student.ID, &result.Student.Name, &result.Student.Class, &result.Student.Gender, &sTelephone, &result.Student.ParentID, &result.Student.CreatedAt, &result.Student.UpdatedAt, &result.Student.DeletedAt,
		&result.Parent.ID, &result.Parent.Name, &result.Parent.Gender, &pTelephone, &result.Parent.Email, &result.Parent.CreatedAt, &result.Parent.UpdatedAt, &result.Parent.DeletedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("student not found")
		}
		return nil, fmt.Errorf("could not get student and parent details: %v", err)
	}

	v, err := strconv.Atoi(pTelephone)
	if err != nil {
		return nil, err
	}

	result.Parent.Telephone = v

	vs, err := strconv.Atoi(sTelephone)
	if err != nil {
		return nil, err
	}

	result.Student.Telephone = vs

	return &result, nil
}

func (sp *studentRepository) UpdateStudent(ctx context.Context, student *domain.Student) error {
	query := `
		UPDATE students
		SET name = $1, class = $2, gender = $3, telephone = $4, parent_id = $5, updated_at = $6
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

// Deleting student would delete the parent too
func (sp *studentRepository) DeleteStudent(ctx context.Context, id int) error {

	var student domain.Student

	query := `
		UPDATE students
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL;
	`
	query2 := `UPDATE parents
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	query3 := `
		SELECT id, name, class, gender, telephone, parent_id, created_at, updated_at, deleted_at
		FROM students
		WHERE id = $1 AND deleted_at IS NULL;
	`

	now := time.Now()

	var telephoneStr string

	// Find student first
	err := sp.db.QueryRow(ctx, query3, id).Scan(
		&student.ID,
		&student.Name,
		&student.Class,
		&student.Gender,
		&telephoneStr,
		&student.ParentID,
		&student.CreatedAt,
		&student.UpdatedAt,
		&student.DeletedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return fmt.Errorf("student not found")
		}
		return fmt.Errorf("could not get student: %v", err)
	}

	// Convert telephone to int
	convertedValue, err := strconv.Atoi(telephoneStr)
	student.Telephone = convertedValue

	if err != nil {
		return fmt.Errorf("invalid telephone format: %v", err)
	}

	// Query delete student
	_, err = sp.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("could not delete student: %v", err)
	}

	// Query delete parent
	_, err = sp.db.Exec(ctx, query2, now, student.ParentID)
	if err != nil {
		return fmt.Errorf("could not delete parent: %v", err)
	}

	return nil
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
