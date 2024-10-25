package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(database *gorm.DB) domain.UserRepo {
	return &userRepository{
		db: database,
	}
}

func (ur *userRepository) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	usernameLowered := strings.ToLower(username)
	err := ur.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", usernameLowered).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("could not find user: %v", err)
		}
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) CreateStaff(ctx context.Context, payload *domain.User) (*domain.User, error) {
	payloadUsernameLowered := strings.ToLower(payload.Username)
	// Check if the user already exists
	var existingUser domain.User
	err := ur.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", payloadUsernameLowered).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("username %s already exists", payloadUsernameLowered)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %v", err)
	}
	payload.Password = string(hashedPassword)

	// Save the new user
	payload.Username = payloadUsernameLowered
	err = ur.db.WithContext(ctx).Create(payload).Error
	if err != nil {
		return nil, fmt.Errorf("could not create user: %v", err)
	}

	return payload, nil
}

func (ur *userRepository) GetAllStaff(ctx context.Context) (*[]domain.SafeStaffData, error) {
	var users []domain.SafeStaffData
	err := ur.db.WithContext(ctx).Model(&domain.User{}).Where("deleted_at IS NULL").Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("could not get all staff: %v", err)
	}

	// Filter out admin users
	filteredUsers := []domain.SafeStaffData{}
	for _, user := range users {
		if user.Role != "admin" {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return &filteredUsers, nil
}

func (ur *userRepository) DeleteStaff(ctx context.Context, id int) error {
	var user domain.User
	err := ur.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("staff not found")
		}
		return fmt.Errorf("could not get staff details: %v", err)
	}

	// Ensure the user is not an admin
	if user.Role == "admin" {
		return fmt.Errorf("could not delete staff")
	}

	// Soft delete the staff
	now := time.Now()
	user.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	err = ur.db.WithContext(ctx).Save(&user).Error
	if err != nil {
		return fmt.Errorf("could not delete staff: %v", err)
	}

	return nil
}

func (ur *userRepository) UpdateStaff(ctx context.Context, id int, payload *domain.User) error {
	usernameLowered := strings.ToLower(payload.Username)

	// Check if the username already exists for another user
	var existingUser domain.User
	err := ur.db.WithContext(ctx).Where("username = ? AND user_id != ? AND deleted_at IS NULL", usernameLowered, id).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("username %s already exists", usernameLowered)
	}

	updateUser := domain.User{
		Username:  usernameLowered,
		Role:      "staff",
		UpdatedAt: time.Now(),
	}

	// Hash the password if it has been updated
	if payload.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("could not hash password: %v", err)
		}
		updateUser.Password = string(hashedPassword)
	}

	// Update the user details
	err = ur.db.WithContext(ctx).Model(&domain.User{}).
		Where("user_id = ? AND deleted_at IS NULL", id).
		Updates(&updateUser).Error

	if err != nil {
		// Check for unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("username %s already exists", usernameLowered)
		}
		return fmt.Errorf("could not update staff: %v", err)
	}

	return nil
}

func (ur *userRepository) GetStaffDetail(ctx context.Context, id int) (*domain.SafeStaffData, error) {
	var user domain.User
	err := ur.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("staff not found")
		}
		return nil, fmt.Errorf("could not get staff details: %v", err)
	}

	if user.Role != "staff" {
		return nil, fmt.Errorf("staff not found")
	}

	safeData := domain.SafeStaffData{
		UserID:    user.UserID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: &user.DeletedAt.Time,
	}

	return &safeData, nil
}

func (ur *userRepository) CreateSubject(ctx context.Context, subject *domain.Subject) error {
	nameLowered := strings.ToLower(subject.Name)

	var existingUser domain.Subject
	err := ur.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", nameLowered).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("subject with %s name already exists", nameLowered)
	}

	subject.Name = nameLowered
	err = ur.db.WithContext(ctx).Create(subject).Error
	if err != nil {
		return fmt.Errorf("could not create subject: %v", err)
	}

	return nil
}

func (ur *userRepository) CreateSubjectBulk(ctx context.Context, subjects *[]domain.Subject) (*[]string, error) {
	var errList []string

	for _, subject := range *subjects {
		loweredName := strings.ToLower(subject.Name)

		var existingSubject domain.Subject
		err := ur.db.WithContext(ctx).Where("LOWER(name) = ?", loweredName).First(&existingSubject).Error

		if err == nil {
			errList = append(errList, fmt.Sprintf("Subject with name %s already exist", subject.Name))
		} else if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	if len(errList) > 0 {
		return &errList, nil
	}

	if err := ur.db.WithContext(ctx).Create(subjects).Error; err != nil {
		return nil, err
	}

	return nil, nil
}

func (ur *userRepository) GetAllSubject(ctx context.Context) (*[]domain.Subject, error) {
	var subjects []domain.Subject
	err := ur.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&subjects).Error
	if err != nil {
		return nil, err
	}
	return &subjects, nil
}

func (ur *userRepository) UpdateSubject(ctx context.Context, id int, newSubjectData *domain.Subject) error {
	nameLowered := strings.ToLower(newSubjectData.Name)

	var existingSubject domain.Subject
	err := ur.db.WithContext(ctx).Where("name = ? AND subject_id != ? AND deleted_at IS NULL", nameLowered, id).First(&existingSubject).Error
	if err == nil {
		return fmt.Errorf("subject with name %s already exists", nameLowered)
	}

	newSubjectData.Name = nameLowered

	err = ur.db.WithContext(ctx).Model(&domain.Subject{}).
		Where("subject_id = ? AND deleted_at IS NULL", id).
		Updates(&newSubjectData).Error
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("subject with name %s already exists", nameLowered)
		}
		return fmt.Errorf("could not update staff: %v", err)
	}

	return nil

}

func (ur *userRepository) DeleteSubject(ctx context.Context, id int) error {
	var subject domain.Subject
	err := ur.db.WithContext(ctx).Where("subject_id = ? AND deleted_at IS NULL", id).First(&subject).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("subject not found")
		}
		return fmt.Errorf("could not get subject details: %v", err)
	}

	now := time.Now()
	subject.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
	err = ur.db.WithContext(ctx).Save(&subject).Error
	if err != nil {
		return fmt.Errorf("could not delete subject: %v", err)
	}

	return nil
}

// func (ur *userRepository) GetlAllClass(ctx context.Context) (*[]domain.Class, error) {
// 	var classess []domain.Class
// 	err := ur.db.WithContext(ctx).Model(&domain.Class{}).Where("deleted_at IS NULL").Find(&classess).Error
// 	if err != nil {
// 		return nil, fmt.Errorf("could not get all class: %v", err)
// 	}

// 	return &classess, nil
// }

// func (ur *userRepository) CreateClass(ctx context.Context, data *domain.Class) error {
// 	err := ur.db.WithContext(ctx).Create(&data).Error
// 	if err != nil {
// 		return fmt.Errorf("could not create class : %v", err)
// 	}

// 	return nil
// }

// func (ur *userRepository) DeleteClass(ctx context.Context, id int) error {
// 	db := ur.db.WithContext(ctx)

// 	if err := db.Model(&domain.Class{}).Where("class_id = ?", id).Update("deleted_at", time.Now()).Error; err != nil {
// 		return fmt.Errorf("failed to soft delete class with ID %d: %w", id, err)
// 	}

// 	return nil
// }
