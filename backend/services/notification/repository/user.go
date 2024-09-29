package repository

import (
	"context"
	"fmt"
	"notification/domain"

	"github.com/jackc/pgx/v5/pgxpool"
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
		WHERE username = $1;
	`

	var user domain.User
	err := ur.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}

	return &user, nil
}

func (ur *userRepository) CreateStaff(ctx context.Context, payload *domain.User) (*domain.User, error) {
	query := `
		SELECT id, username, role
		FROM users
		WHERE username = $1;
	`

	var user domain.User
	err := ur.db.QueryRow(ctx, query, payload.Username).Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		return nil, fmt.Errorf("could not find user: %v", err)
	}

	return &user, nil
}
