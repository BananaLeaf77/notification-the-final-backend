package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type studentParentRepository struct {
	db *pgxpool.Pool
}

func NewStudentParentRepository(database *pgxpool.Pool) domain.StudentParentRepo {
	return &studentParentRepository{
		db: database,
	}
}

func (spr *studentParentRepository) CreateStudentAndParent(ctx context.Context, req *domain.StudentAndParent) error {
	studTelephone := strconv.Itoa(req.Student.Telephone)
	parentTelephone := strconv.Itoa(req.Parent.Telephone)

	// Check if the student telephone already exists
	checkStudentTelephoneQuery := `SELECT id FROM students WHERE telephone = $1 AND deleted_at IS NULL;`
	var existingStudentID int
	err := spr.db.QueryRow(ctx, checkStudentTelephoneQuery, studTelephone).Scan(&existingStudentID)
	if err == nil {
		// Student with this telephone already exists
		return fmt.Errorf("student with telephone %s already exists", studTelephone)
	} else if err != pgx.ErrNoRows {
		return fmt.Errorf("error checking student telephone: %v", err)
	}

	// Check if the parent telephone already exists
	checkParentTelephoneQuery := `SELECT id FROM parents WHERE telephone = $1 AND deleted_at IS NULL;`
	var existingParentID int
	err = spr.db.QueryRow(ctx, checkParentTelephoneQuery, parentTelephone).Scan(&existingParentID)
	if err == nil {
		// Parent with this telephone already exists
		return fmt.Errorf("parent with telephone %s already exists", parentTelephone)
	} else if err != pgx.ErrNoRows {
		return fmt.Errorf("error checking parent telephone: %v", err)
	}

	// Check if the parent email already exists
	checkParentEmailQuery := `SELECT id FROM parents WHERE email = $1 AND deleted_at IS NULL;`
	err = spr.db.QueryRow(ctx, checkParentEmailQuery, req.Parent.Email).Scan(&existingParentID)
	if err == nil {
		// Parent with this email already exists
		return fmt.Errorf("parent with email %s already exists", *req.Parent.Email)
	} else if err != pgx.ErrNoRows {
		return fmt.Errorf("error checking parent email: %v", err)
	}

	// If all checks pass, proceed with the transaction
	tx, err := spr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Insert parent
	parentInsertQuery := `
		INSERT INTO parents (name, gender, telephone, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	now := time.Now()
	var parentID int
	err = tx.QueryRow(ctx, parentInsertQuery, req.Parent.Name, req.Parent.Gender, parentTelephone, req.Parent.Email, now, now).Scan(&parentID)
	if err != nil {
		return fmt.Errorf("could not insert parent: %v", err)
	}

	req.Parent.ID = parentID
	req.Parent.CreatedAt = now
	req.Parent.UpdatedAt = now

	// Insert student
	studentInsertQuery := `
		INSERT INTO students (name, class, gender, telephone, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	var studentID int
	err = tx.QueryRow(ctx, studentInsertQuery, req.Student.Name, req.Student.Class, req.Student.Gender, studTelephone, parentID, now, now).Scan(&studentID)
	if err != nil {
		return fmt.Errorf("could not insert student: %v", err)
	}

	req.Student.ID = studentID
	req.Student.CreatedAt = now
	req.Student.UpdatedAt = now

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (spr *studentParentRepository) ImportCSV(ctx context.Context, payload *[]domain.StudentAndParent) (*[]string, error) {

	var parentTeleponeSTR string
	var studentTeleponeSTR string
	var duplicateMessages []string

	now := time.Now()

	// Prepare statements for inserting parent and student data
	parentInsertQuery := `
		INSERT INTO parents (name, gender, telephone, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`
	studentInsertQuery := `
		INSERT INTO students (name, class, gender, telephone, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`

	// Queries to check if parent or student already exists
	checkParentExistsByTelephoneQuery := `SELECT id FROM parents WHERE telephone = $1 AND deleted_at IS NULL;`
	checkParentExistsByEmailQuery := `SELECT id FROM parents WHERE email = $1 AND deleted_at IS NULL;`
	checkStudentExistsQuery := `SELECT id FROM students WHERE telephone = $1 AND deleted_at IS NULL;`

	for index, record := range *payload {
		var parentExistsID int

		parentTeleponeSTR = fmt.Sprintf("0%s", strconv.Itoa(record.Parent.Telephone))
		studentTeleponeSTR = fmt.Sprintf("0%s", strconv.Itoa(record.Student.Telephone))

		// Check if parent already exists by telephone
		err := spr.db.QueryRow(ctx, checkParentExistsByTelephoneQuery, parentTeleponeSTR).Scan(&parentExistsID)
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with telephone %s already exists", index+1, parentTeleponeSTR))
		} else if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("row %d: error checking if parent exists by telephone: %v", index+1, err)
		}

		// Check if parent already exists by email
		err = spr.db.QueryRow(ctx, checkParentExistsByEmailQuery, record.Parent.Email).Scan(&parentExistsID)
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with email %s already exists", index+1, *record.Parent.Email))
		} else if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("row %d: error checking if parent exists by email: %v", index+1, err)
		}

		// Check if student already exists
		var studentExistsID int
		err = spr.db.QueryRow(ctx, checkStudentExistsQuery, studentTeleponeSTR).Scan(&studentExistsID)
		if err == nil {
			// Student already exists, add a message to duplicates and continue
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with telephone %s already exists", index+1, studentTeleponeSTR))

		} else if err != pgx.ErrNoRows {
			// If there's an error other than "no rows", we should handle it
			return nil, fmt.Errorf("row %d: error checking if student exists: %v", index+1, err)
		}
	}

	// check jika panjang  var duplikat msg lebih dari 0 (ada)
	if len(duplicateMessages) > 0 {
		return &duplicateMessages, nil
	}

	for index, record := range *payload {

		parentTeleponeSTR = fmt.Sprintf("0%s", strconv.Itoa(record.Parent.Telephone))
		studentTeleponeSTR = fmt.Sprintf("0%s", strconv.Itoa(record.Student.Telephone))

		// Insert parent
		var parentID int

		err := spr.db.QueryRow(ctx, parentInsertQuery, record.Parent.Name, record.Parent.Gender, parentTeleponeSTR, record.Parent.Email, now, now).Scan(&parentID)
		if err != nil {
			return nil, fmt.Errorf("row %d: could not insert parent: %v", index+1, err)
		}

		// Update ParentID in record
		record.Parent.ID = parentID
		record.Parent.CreatedAt = now
		record.Parent.UpdatedAt = now

		// Insert student with the retrieved ParentID
		var studentID int

		err = spr.db.QueryRow(ctx, studentInsertQuery, record.Student.Name, record.Student.Class, record.Student.Gender, studentTeleponeSTR, parentID, now, now).Scan(&studentID)
		if err != nil {
			return nil, fmt.Errorf("row %d: could not insert student: %v", index+1, err)
		}

		// Update StudentID in record
		record.Student.ID = studentID
		record.Student.CreatedAt = now
		record.Student.UpdatedAt = now

	}

	return &duplicateMessages, nil
}

func (r *studentParentRepository) UpdateStudentAndParent(ctx context.Context, id int, payload *domain.StudentAndParent) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Prepare formatted telephone numbers
	studentTelephone := fmt.Sprintf("0%s", strconv.Itoa(payload.Student.Telephone))
	parentTelephone := fmt.Sprintf("0%s", strconv.Itoa(payload.Parent.Telephone))

	// Check if the student's telephone already exists, excluding the current student
	checkStudentTelephoneQuery := `
		SELECT COUNT(1)
		FROM students
		WHERE telephone = $1 AND id != $2 AND deleted_at IS NULL;
	`
	var studentCount int
	err = r.db.QueryRow(ctx, checkStudentTelephoneQuery, studentTelephone, id).Scan(&studentCount)
	if err != nil {
		return fmt.Errorf("error checking student telephone: %v", err)
	}
	if studentCount > 0 {
		return fmt.Errorf("telephone number %s already exists for another student", studentTelephone)
	}

	// Check if the parent's telephone already exists, excluding the current parent
	checkParentTelephoneQuery := `
		SELECT COUNT(1)
		FROM parents
		WHERE telephone = $1 AND id != $2 AND deleted_at IS NULL;
	`
	var parentCount int
	err = r.db.QueryRow(ctx, checkParentTelephoneQuery, parentTelephone, payload.Student.ParentID).Scan(&parentCount)
	if err != nil {
		return fmt.Errorf("error checking parent telephone: %v", err)
	}
	if parentCount > 0 {
		return fmt.Errorf("telephone number %s already exists for another parent", parentTelephone)
	}

	// Check if the parent's email already exists, excluding the current parent
	if payload.Parent.Email != nil {
		checkParentEmailQuery := `
			SELECT COUNT(1)
			FROM parents
			WHERE email = $1 AND id != $2 AND deleted_at IS NULL;
		`
		var emailCount int
		err = r.db.QueryRow(ctx, checkParentEmailQuery, *payload.Parent.Email, payload.Student.ParentID).Scan(&emailCount)
		if err != nil {
			return fmt.Errorf("error checking parent email: %v", err)
		}
		if emailCount > 0 {
			return fmt.Errorf("email %s already exists for another parent", *payload.Parent.Email)
		}
	}

	// Update the student
	studentUpdateQuery := `
		UPDATE students
		SET name = $1, class = $2, gender = $3, telephone = $4, updated_at = $5
		WHERE id = $6;
	`
	_, err = tx.Exec(ctx, studentUpdateQuery, payload.Student.Name, payload.Student.Class, payload.Student.Gender, studentTelephone, time.Now(), id)
	if err != nil {
		return fmt.Errorf("could not update student: %v", err)
	}

	// Update the parent
	parentUpdateQuery := `
		UPDATE parents
		SET name = $1, gender = $2, telephone = $3, email = $4, updated_at = $5
		WHERE id = $6;
	`
	_, err = tx.Exec(ctx, parentUpdateQuery, payload.Parent.Name, payload.Parent.Gender, parentTelephone, payload.Parent.Email, time.Now(), payload.Student.ParentID)
	if err != nil {
		return fmt.Errorf("could not update parent: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (r *studentParentRepository) GetStudentDetailByID(ctx context.Context, id int) (*domain.StudentAndParent, error) {
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

	err := r.db.QueryRow(ctx, query, id).Scan(
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

func (sp *studentParentRepository) DeleteStudentAndParent(ctx context.Context, id int) error {

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
