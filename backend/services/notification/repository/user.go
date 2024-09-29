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
