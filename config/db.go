package config

import (
	"fmt"
	"notification/domain"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetDatabaseURL builds the database connection string.
func GetDatabaseURL() string {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_DATABASE"))
	return dsn
}

// BootDB initializes the database connection and runs migrations.
func BootDB() (*gorm.DB, error) {
	url := GetDatabaseURL()
	var err error

	db, err = gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the models
	if err := autoMigrate(db); err != nil {
		return db, err
	}

	fmt.Println("DB initialized")
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	// Pastikan ENUM sudah dibuat sebelum digunakan
	if err := db.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_enum') THEN
			CREATE TYPE gender_enum AS ENUM ('male', 'female');
		END IF;
	END $$`).Error; err != nil {
		return fmt.Errorf("failed to create gender ENUM: %w", err)
	}

	if err := db.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_enum') THEN
			CREATE TYPE role_enum AS ENUM ('admin', 'staff');
		END IF;
	END $$`).Error; err != nil {
		return fmt.Errorf("failed to create role ENUM: %w", err)
	}

	// Migrasi tabel yang tidak memiliki foreign key lebih dulu
	if err := db.AutoMigrate(
		&domain.Parent{},
		&domain.Student{},
		&domain.User{},
		&domain.Subject{},
	); err != nil {
		return fmt.Errorf("failed to migrate base tables: %w", err)
	}

	// Migrasi tabel yang memiliki foreign key
	if err := db.AutoMigrate(
		&domain.TestScore{},
		&domain.AttendanceNotificationHistory{},
		&domain.ParentDataChangeRequest{},
	); err != nil {
		return fmt.Errorf("failed to migrate relational tables: %w", err)
	}

	var existingAdmin domain.User
	err := db.Where("role = 'admin' AND deleted_at IS NULL").First(&existingAdmin).Error
	if err != nil {
		fmt.Println("Creating default admin account....")
		adminUsername := os.Getenv("ADMIN_USERNAME")
		adminName := os.Getenv("ADMIN_NAME")
		adminPassword := os.Getenv("ADMIN_PASSWORD")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("could not hash password: %v", err)
		}

		now := time.Now()
		admin := domain.User{
			Username:  adminUsername,
			Name:      adminName,
			Password:  string(hashedPassword),
			Role:      "admin",
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = db.Create(&admin).Error
		if err != nil {
			return err
		}
		fmt.Println("Admin account created")
	}

	return nil
}
