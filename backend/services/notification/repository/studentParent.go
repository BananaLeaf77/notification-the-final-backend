package repository

import (
	"context"
	"fmt"
	"notification/domain"
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
	err = tx.QueryRow(ctx, parentInsertQuery, req.Parent.Name, req.Parent.Gender, req.Parent.Telephone, req.Parent.Email, now, now).Scan(&parentID)
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
	err = tx.QueryRow(ctx, studentInsertQuery, req.Student.Name, req.Student.Class, req.Student.Gender, req.Student.Telephone, parentID, now, now).Scan(&studentID)
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

func (spr *studentParentRepository) GetStudentAndParent(ctx context.Context, studentID string) (*domain.StudentAndParent, error) {
	return nil, nil
}

func (spr *studentParentRepository) ImportCSV(ctx context.Context, payload *[]domain.StudentAndParent) (*[]string, error) {
	tx, err := spr.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %v", err)
	}
	defer tx.Rollback(ctx) // Ensure rollback on error

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
	checkParentExistsQuery := `
		SELECT id FROM parents WHERE (telephone = $1 OR email = $2) AND deleted_at IS NULL;
	`
	checkStudentExistsQuery := `
		SELECT id FROM students WHERE telephone = $1 AND deleted_at IS NULL;
	`

	// Slice to store duplicate messages
	var duplicateMessages []string

	for index, record := range *payload {
		// Check if parent already exists
		var parentExistsID int
		err := tx.QueryRow(ctx, checkParentExistsQuery, record.Parent.Telephone, record.Parent.Email).Scan(&parentExistsID)
		if err == nil {
			// Parent already exists, add a message to duplicates and continue
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with telephone %s or email %s already exists", index+1, record.Parent.Telephone, record.Parent.Email))
			continue
		} else if err != pgx.ErrNoRows {
			// If there's an error other than "no rows", we should handle it
			return nil, fmt.Errorf("row %d: error checking if parent exists: %v", index+1, err)
		}

		// Check if student already exists
		var studentExistsID int
		err = tx.QueryRow(ctx, checkStudentExistsQuery, record.Student.Telephone).Scan(&studentExistsID)
		if err == nil {
			// Student already exists, add a message to duplicates and continue
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with telephone %s already exists", index+1, record.Student.Telephone))
			continue
		} else if err != pgx.ErrNoRows {
			// If there's an error other than "no rows", we should handle it
			return nil, fmt.Errorf("row %d: error checking if student exists: %v", index+1, err)
		}

		// Insert parent
		var parentID int
		err = tx.QueryRow(ctx, parentInsertQuery, record.Parent.Name, record.Parent.Gender, record.Parent.Telephone, record.Parent.Email, now, now).Scan(&parentID)
		if err != nil {
			return nil, fmt.Errorf("row %d: could not insert parent: %v", index+1, err)
		}

		// Update ParentID in record
		record.Parent.ID = parentID
		record.Parent.CreatedAt = now
		record.Parent.UpdatedAt = now

		// Insert student with the retrieved ParentID
		var studentID int
		err = tx.QueryRow(ctx, studentInsertQuery, record.Student.Name, record.Student.Class, record.Student.Gender, record.Student.Telephone, parentID, now, now).Scan(&studentID)
		if err != nil {
			return nil, fmt.Errorf("row %d: could not insert student: %v", index+1, err)
		}

		// Update StudentID in record
		record.Student.ID = studentID
		record.Student.CreatedAt = now
		record.Student.UpdatedAt = now
	}

	// If there are any successful inserts, commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("could not commit transaction: %v", err)
	}

	return &duplicateMessages, nil
}

