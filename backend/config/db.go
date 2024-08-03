package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var sqlDB *sql.DB

func GetDatabaseURL() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_DATABASE"))
	return dsn
}

func BootDB() (*sql.DB, error) {
	url := GetDatabaseURL()
	// fmt.Println("Connecting to database with URL:", url)
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if sqlDB == nil {
		sqlDB = db
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}

	err = autoMigrate(sqlDB)
	if err != nil {
		return sqlDB, err
	}

	return sqlDB, nil
}

// fungsi migrate
func autoMigrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    class VARCHAR(10) NOT NULL,
    gender VARCHAR(15) NOT NULL,
    telephone_number BIGINT NOT NULL
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		fmt.Printf("Error executing migration query: %v\n", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
