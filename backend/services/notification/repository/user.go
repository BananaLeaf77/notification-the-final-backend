package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(database *pgxpool.Pool) domain.UserRepo {
	return &userRepository{
		db: database,
	}
}

func (ur *userRepository) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password, role
		FROM users
		WHERE username = $1
		AND deleted_at IS NULL;
	`

	var user domain.User
	err := ur.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}

	return &user, nil
}

func (ur *userRepository) CreateStaff(ctx context.Context, payload *domain.User) (*domain.User, error) {
	checkQuery := `
		SELECT id 
		FROM users
		WHERE username = $1
		AND deleted_at IS NULL;
	`

	var existingUserID int
	err := ur.db.QueryRow(ctx, checkQuery, payload.Username).Scan(&existingUserID)
	if err == nil {
		return nil, fmt.Errorf("username %s already exists", payload.Username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %v", err)
	}

	insertQuery := `
		INSERT INTO users (username, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	now := time.Now()

	err = ur.db.QueryRow(ctx, insertQuery, payload.Username, string(hashedPassword), payload.Role, now, now).Scan(&payload.ID)
	if err != nil {
		fmt.Printf("Error inserting user: %v\n", err)
		return nil, fmt.Errorf("could not create user: %v", err)
	}

	payload.CreatedAt = now
	payload.UpdatedAt = now

	return payload, nil
}

func (ur *userRepository) GetAllStaff(ctx context.Context) (*[]domain.SafeStaffData, error) {
	query := `
		SELECT id, username, role, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL;
	`

	rows, err := ur.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not get all staff: %v", err)
	}
	defer rows.Close()

	var users []domain.SafeStaffData
	for rows.Next() {
		var user domain.SafeStaffData

		err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("could not scan user: %v", err)
		}

		if user.Role != "admin" {
			users = append(users, user)
		}

	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return &users, nil
}

func (ur *userRepository) DeleteStaff(ctx context.Context, id int) error {
	var userHolder domain.SafeStaffData
	now := time.Now()

	query2 := `
		SELECT id, username, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL;
	`

	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL;
	`

	err := ur.db.QueryRow(ctx, query2, id).Scan(&userHolder.ID, &userHolder.Username, &userHolder.Role, &userHolder.CreatedAt, &userHolder.UpdatedAt, &userHolder.DeletedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return fmt.Errorf("student not found")
		}
		return fmt.Errorf("could not get student and parent details: %v", err)
	}

	if userHolder.Role == "admin" {
		return fmt.Errorf("could not delete staff")
	}

	// Execute the query and check the number of rows affected
	result, err := ur.db.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("could not delete staff: %v", err)
	}

	// Check if any rows were affected
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no staff found with id %d", id)
	}

	return nil
}

func (ur *userRepository) UpdateStaff(ctx context.Context, id int, payload *domain.User) error {
	now := time.Now()

	query := `
		UPDATE users
		SET username = $1, password = $2, role = $3, updated_at = $4
		WHERE id = $5;
	`

	_, err := ur.db.Exec(ctx, query, payload.Username, payload.Password, payload.Role, now, id)
	if err != nil {
		return fmt.Errorf("could not update staff: %v", err)
	}

	return nil
}

func (ur *userRepository) GetStaffDetail(ctx context.Context, id int) (*domain.SafeStaffData, error) {
	var valueHolder domain.SafeStaffData
	query := `
		SELECT id, username, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL;
	`

	err := ur.db.QueryRow(ctx, query, id).Scan(&valueHolder.ID, &valueHolder.Username, &valueHolder.Role, &valueHolder.CreatedAt, &valueHolder.UpdatedAt, &valueHolder.DeletedAt)

	if valueHolder.Role != "staff" {
		return nil, fmt.Errorf("staff not found")
	}

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("staff not found")
		}
		return nil, fmt.Errorf("could not get staff details: %v", err)
	}

	return &valueHolder, nil
}
