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

func (r *userRepository) GetSubjectsForTeacher(ctx context.Context, userID int) (*[]domain.Subject, error) {
	var subjects []domain.Subject

	err := r.db.WithContext(ctx).
		Table("subjects").
		Joins("JOIN user_subjects ON user_subjects.subject_subject_id = subjects.subject_id").
		Where("user_subjects.user_user_id = ?", userID).
		Find(&subjects).Error

	if err != nil {
		return nil, err
	}

	return &subjects, nil
}

func (ur *userRepository) CreateStaff(ctx context.Context, payload *domain.User, subjectIDs []int) (*domain.User, error) {
	payloadUsernameLowered := strings.ToLower(payload.Username)

	var existingUser domain.User
	err := ur.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", payloadUsernameLowered).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("username %s already exists", payloadUsernameLowered)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %v", err)
	}
	payload.Password = string(hashedPassword)

	var subjects []domain.Subject
	err = ur.db.WithContext(ctx).Where("subject_id IN ?", subjectIDs).Find(&subjects).Error
	if err != nil {
		return nil, fmt.Errorf("could not assign subjects: %v", err)
	}

	subjectPointers := make([]*domain.Subject, len(subjects))
	for i := range subjects {
		subjectPointers[i] = &subjects[i]
	}

	payload.Teaching = subjectPointers

	// Save the new user
	payload.Username = payloadUsernameLowered
	err = ur.db.WithContext(ctx).Create(payload).Error
	if err != nil {
		return nil, fmt.Errorf("could not create user: %v", err)
	}

	return payload, nil
}

func (ur *userRepository) GetAllStaff(ctx context.Context) (*[]domain.SafeStaffData, error) {
	var users []domain.User
	err := ur.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("could not get all staff: %v", err)
	}

	// Prepare to hold safe staff data
	var safeStaffData []domain.SafeStaffData

	for _, user := range users {
		// Skip admin users
		if user.Role == "admin" {
			continue
		}

		// Fetch subjects associated with the user
		var subjects []domain.Subject
		if err := ur.db.WithContext(ctx).Model(&user).
			Association("Teaching").Find(&subjects); err != nil {
			return nil, fmt.Errorf("could not get subjects for user %d: %v", user.UserID, err)
		}

		// Convert gorm.DeletedAt to *time.Time
		var deletedAt *time.Time
		if user.DeletedAt.Valid {
			deletedAt = &user.DeletedAt.Time
		}

		// Create SafeStaffData instance
		safeStaffData = append(safeStaffData, domain.SafeStaffData{
			UserID:    user.UserID,
			Username:  user.Username,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: deletedAt,
			Teaching:  subjects,
		})
	}

	return &safeStaffData, nil
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

func (ur *userRepository) UpdateStaff(ctx context.Context, id int, payload *domain.User, subjectIDs []int) error {
	usernameLowered := strings.ToLower(payload.Username)

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

	err = ur.db.WithContext(ctx).Model(&domain.User{}).
		Where("user_id = ? AND deleted_at IS NULL", id).
		Updates(&updateUser).Error

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("username %s already exists", usernameLowered)
		}
		return fmt.Errorf("could not update staff: %v", err)
	}

	if len(subjectIDs) > 0 {
		var user domain.User
		if err := ur.db.WithContext(ctx).First(&user, id).Error; err != nil {
			return fmt.Errorf("could not find user: %v", err)
		}

		if err := ur.db.WithContext(ctx).Model(&user).Association("Teaching").Clear(); err != nil {
			return fmt.Errorf("could not clear existing subjects: %v", err)
		}

		var subjects []domain.Subject
		if err := ur.db.WithContext(ctx).Where("subject_id IN ?", subjectIDs).Find(&subjects).Error; err != nil {
			return fmt.Errorf("could not find new subjects: %v", err)
		}

		subjectPointers := make([]*domain.Subject, len(subjects))
		for i := range subjects {
			subjectPointers[i] = &subjects[i]
		}

		if err := ur.db.WithContext(ctx).Model(&user).Association("Teaching").Replace(subjectPointers); err != nil {
			return fmt.Errorf("could not update subjects: %v", err)
		}
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
