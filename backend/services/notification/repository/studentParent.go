package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"time"

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

func (spr *studentParentRepository) GetStudentAndParent(ctx context.Context, studentID string) (*domain.StudentAndParent, error){

}