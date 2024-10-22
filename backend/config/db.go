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
	// Create ENUM type for gender if not exists
	if err := db.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_enum') THEN
			CREATE TYPE gender_enum AS ENUM ('male', 'female');
		END IF;
	END $$`).Error; err != nil {
		fmt.Printf("Error creating gender ENUM type: %v\n", err)
		return fmt.Errorf("failed to create gender ENUM: %w", err)
	}

	if err := db.Exec(`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_enum') THEN
			CREATE TYPE role_enum AS ENUM ('admin', 'staff');
		END IF;
	END $$`).Error; err != nil {
		fmt.Printf("Error creating role ENUM type: %v\n", err)
		return fmt.Errorf("failed to create role ENUM: %w", err)
	}

	theParent := domain.Parent{}
	theStudent := domain.Student{}
	theUser := domain.User{}
	theNotificationHistory := domain.AttendanceNotificationHistory{}
	theDataChangeRequest := domain.DataChangeRequest{}

	if err := db.AutoMigrate(&theParent, &theStudent, &theUser, &theDataChangeRequest); err != nil {
		return fmt.Errorf("failed to run initial migrations: %w", err)
	}

	if err := db.AutoMigrate(&theNotificationHistory); err != nil {
		return fmt.Errorf("failed to run notification history migration: %w", err)
	}

	var existingAdmin domain.User
	err := db.Where("role = 'admin' AND deleted_at IS NULL").First(&existingAdmin).Error
	if err != nil {
		fmt.Println("Creating default admin account....")
		adminUsername := os.Getenv("ADMIN_USERNAME")
		adminPassword := os.Getenv("ADMIN_PASSWORD")

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("could not hash password: %v", err)
		}

		now := time.Now()
		admin := domain.User{
			Username:  adminUsername,
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
