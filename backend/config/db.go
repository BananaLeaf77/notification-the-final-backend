package config

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pgxPool *pgxpool.Pool

func GetDatabaseURL() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_DATABASE"))
	return dsn
}

func BootDB() (*pgxpool.Pool, error) {
	url := GetDatabaseURL()
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if pgxPool == nil {
		pgxPool = dbPool
	}

	err = pgxPool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	err = autoMigrate(pgxPool)
	if err != nil {
		return pgxPool, err
	}

	return pgxPool, nil
}

func autoMigrate(pool *pgxpool.Pool) error {
	createParentsTableQuery := `
	CREATE TABLE IF NOT EXISTS parents (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		gender VARCHAR(15) NOT NULL,
		telephone_number BIGINT NOT NULL,
		email VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE
	);
	`
	_, err := pool.Exec(context.Background(), createParentsTableQuery)
	if err != nil {
		fmt.Printf("Error executing parents table migration query: %v\n", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	createStudentsTableQuery := `
	CREATE TABLE IF NOT EXISTS students (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		class VARCHAR(10) NOT NULL,
		gender VARCHAR(15) NOT NULL,
		telephone_number BIGINT NOT NULL,
		parent_id INTEGER,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP WITH TIME ZONE,
		CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES parents(id)
	);
	`
	_, err = pool.Exec(context.Background(), createStudentsTableQuery)
	if err != nil {
		fmt.Printf("Error executing students table migration query: %v\n", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
