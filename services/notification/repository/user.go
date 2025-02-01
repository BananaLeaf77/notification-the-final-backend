package repository

import (
	"context"
	"errors"
	"fmt"
	"notification/config"
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

func (ur *userRepository) GetAllTestScoresBySubjectID(ctx context.Context, subjectID int) (*[]domain.TestScore, error) {
	// Get the subject first
	var subject domain.Subject
	var testScores []domain.TestScore
	var students []domain.Student

	// Fetch the subject
	err := ur.db.WithContext(ctx).Where("subject_id = ?", subjectID).First(&subject).Error
	if err != nil {
		return nil, err
	}

	// Fetch all students with matching grade
	err = ur.db.WithContext(ctx).Where("grade = ?", subject.Grade).Find(&students).Error
	if err != nil {
		return nil, err
	}

	// Fetch test scores for the subject
	err = ur.db.WithContext(ctx).
		Preload("Student").
		Preload("Subject").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id", "username", "name", "role", "created_at", "updated_at", "deleted_at")
		}).
		Where("subject_id = ? AND deleted_at IS NULL", subjectID).
		Find(&testScores).Error
	if err != nil {
		return nil, err
	}

	// Filter test scores to include only those with active students
	var validTestScores []domain.TestScore
	for _, testScore := range testScores {
		// Check if the associated student is active
		var student domain.Student
		studentCheckErr := ur.db.WithContext(ctx).
			Where("student_id = ?", testScore.StudentID).
			First(&student).Error

		// Include the test score only if the student is active
		if studentCheckErr == nil {
			validTestScores = append(validTestScores, testScore)
		}
	}

	// Create a map of students with valid test scores
	validStudentIDs := make(map[int]struct{})
	for _, testScore := range validTestScores {
		validStudentIDs[testScore.StudentID] = struct{}{}
	}

	// Add students without test scores if they are active
	for _, student := range students {
		if _, exists := validStudentIDs[student.StudentID]; !exists {
			validTestScores = append(validTestScores, domain.TestScore{
				StudentID: student.StudentID,
				Student:   student,
				Score:     floatPointer(0), // Set Score to 0
			})
		}
	}

	return &validTestScores, nil
}

func floatPointer(f float64) *float64 {
	return &f
}

func (ur *userRepository) GetAllTestScores(ctx context.Context) (*[]domain.TestScore, error) {
	var testScores []domain.TestScore
	err := ur.db.WithContext(ctx).Preload("Student").Preload("User").Preload("Subject").Where("deleted_at IS NULL").Find(&testScores).Error
	if err != nil {
		return nil, err
	}

	return &testScores, nil
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

func (r *userRepository) InputTestScores(ctx context.Context, teacherID int, testScores *domain.InputTestScorePayload) error {
	tx := r.db.WithContext(ctx).Begin()

	var userDetail domain.User
	err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at is NULL", teacherID).First(&userDetail).Error
	if err != nil {
		return fmt.Errorf("user with id %d not found", teacherID)
	}

	var subject domain.Subject
	if err := tx.Where("subject_id = ?", testScores.SubjectID).First(&subject).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("subject ID %d does not exist", testScores.SubjectID)
	}

	for _, individual := range testScores.StudentTestScore {
		var student domain.Student
		if err := tx.Where("student_id = ?", individual.StudentID).First(&student).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("student ID %d does not exist", individual.StudentID)
		}

		// Authorization check for non-admin users
		if userDetail.Role != "admin" {
			var count int64
			err := tx.Table("user_subjects").
				Where("user_user_id = ? AND subject_subject_id = ?", teacherID, testScores.SubjectID).
				Count(&count).Error

			if err != nil || count == 0 {
				tx.Rollback()
				return fmt.Errorf("user is not authorized to input scores for subject ID %d", testScores.SubjectID)
			}
		}

		// Check if a test individual already exists for this student and subject (ignore teacher)
		var existingScore domain.TestScore
		err := tx.Where("student_id = ? AND subject_id = ?", individual.StudentID, testScores.SubjectID).
			First(&existingScore).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}

		if existingScore.TestScoreID > 0 {
			// Update the existing individual
			existingScore.Score = individual.TestScore
			existingScore.UserID = teacherID // Optionally update the teacher ID to the new one
			if err := tx.Save(&existingScore).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			// Create a new test individual record
			newScore := domain.TestScore{
				StudentID: individual.StudentID,
				SubjectID: testScores.SubjectID,
				UserID:    teacherID,
				Score:     individual.TestScore,
			}
			config.PrintStruct(newScore)
			if err := tx.Create(&newScore).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

func (r *userRepository) GetSubjectsForTeacher(ctx context.Context, userID int) (*domain.SafeStaffData, error) {
	var subjects []domain.Subject
	var user domain.User
	var safeData domain.SafeStaffData

	// Retrieve the user and preload their teaching subjects
	err := r.db.WithContext(ctx).Preload("Teaching").Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("user with id %d not found", userID)
	}

	// If the user is an admin, retrieve all subjects
	if user.Role == "admin" {
		err = r.db.WithContext(ctx).Find(&subjects).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get all subjects: %v", err)
		}

		// Initialize Teaching slice in safeData
		safeData.Teaching = make([]domain.Subject, len(subjects))

		// Use copy instead of the loop to copy the subjects into safeData.Teaching
		copy(safeData.Teaching, subjects)

		safeData.UserID = user.UserID
		safeData.Name = user.Name
		safeData.Role = user.Role

		return &safeData, nil
	}

	safeData.UserID = user.UserID
	safeData.Name = user.Name
	safeData.Username = user.Username
	safeData.Role = user.Role

	safeData.Teaching = make([]domain.Subject, len(user.Teaching))
	for i, subject := range user.Teaching {
		safeData.Teaching[i] = *subject
	}

	return &safeData, nil
}

func (ur *userRepository) CreateStaff(ctx context.Context, payload *domain.User) (*domain.User, error) {
	payloadUsernameLowered := strings.ToLower(payload.Username)
	// Check if username already exists
	var existingUser domain.User
	err := ur.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", payloadUsernameLowered).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("username %s already exists", payloadUsernameLowered)
	}

	err = ur.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", payload.Name).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("name %s already exists", payload.Name)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %v", err)
	}
	payload.Password = string(hashedPassword)

	// Save the new user (this creates a user record in the user table)
	payload.Username = payloadUsernameLowered
	payload.Role = "staff"
	err = ur.db.WithContext(ctx).Create(payload).Error
	if err != nil {
		return nil, fmt.Errorf("could not create user: %v", err)
	}

	return payload, nil
}

func (ur *userRepository) GetAllStaff(ctx context.Context) (*[]domain.SafeStaffData, error) {
	var users []domain.User
	err := ur.db.WithContext(ctx).Preload("Teaching").Where("deleted_at IS NULL").Find(&users).Error
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

		// Convert []*domain.Subject to []domain.Subject
		teaching := make([]domain.Subject, len(user.Teaching))
		for i, subject := range user.Teaching {
			teaching[i] = *subject
		}

		safeStaffData = append(safeStaffData, domain.SafeStaffData{
			UserID:    user.UserID,
			Username:  user.Username,
			Name:      user.Name,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			DeletedAt: user.DeletedAt,
			Teaching:  teaching,
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
		return fmt.Errorf("could not delete staff: user is an admin")
	}

	// Soft delete the staff
	now := time.Now()
	user.DeletedAt = &now // Assign the current time to mark as deleted
	err = ur.db.WithContext(ctx).Save(&user).Error
	if err != nil {
		return fmt.Errorf("could not delete staff: %v", err)
	}

	return nil
}

func (ur *userRepository) UpdateStaff(ctx context.Context, id int, payload *domain.User, subjectIDs []int) error {
	usernameLowered := strings.ToLower(payload.Username)
	var foundUser domain.User
	err := ur.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", id).First(&foundUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("staff not found")
		}
		return fmt.Errorf("could not get staff details: %v", err)
	}

	if foundUser.Role == "admin" {
		return fmt.Errorf("cant modify admin")
	}

	var existingUser domain.User
	err = ur.db.WithContext(ctx).Where("username = ? AND user_id != ? AND deleted_at IS NULL", usernameLowered, id).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("username %s already exists", usernameLowered)
	}

	err = ur.db.WithContext(ctx).Where("name = ? AND user_id != ? AND deleted_at IS NULL", payload.Name, id).First(&existingUser).Error
	if err == nil {
		return fmt.Errorf("name %s already exists", payload.Name)
	}

	updateUser := domain.User{
		Username:  usernameLowered,
		Name:      payload.Name,
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

		if len(subjects) == 0 {
			return fmt.Errorf("no subjects found for the given IDs")
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

func (ur *userRepository) ShowProfile(ctx context.Context, uID int) (*domain.SafeStaffData, error) {
	var user domain.User
	var safeUser domain.SafeStaffData

	err := ur.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", uID).First(&user).Error
	if err != nil {
		return nil, err
	}
	safeUser.Username = user.Username
	safeUser.UserID = user.UserID
	safeUser.Name = user.Name
	safeUser.Role = user.Role

	return &safeUser, nil
}

func (ur *userRepository) GetAdminByAdmin(ctx context.Context) (*domain.SafeStaffData, error) {
	var subjects []domain.Subject
	var admin domain.User
	var adminSafeData domain.SafeStaffData

	err := ur.db.WithContext(ctx).Where("user_id = 1 AND deleted_at IS NULL").First(&admin).Error
	if err != nil {
		return nil, err
	}
	adminSafeData.UserID = admin.UserID
	adminSafeData.Name = admin.Name
	adminSafeData.Role = admin.Role

	if err := ur.db.WithContext(ctx).Find(&subjects).Error; err != nil {
		return nil, fmt.Errorf("could not fetch subjects for admin: %w", err)
	}

	adminSafeData.Teaching = subjects

	return &adminSafeData, nil
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

	var subjects []domain.Subject
	if err := ur.db.WithContext(ctx).Model(&user).
		Association("Teaching").Find(&subjects); err != nil {
		return nil, fmt.Errorf("could not get subjects for user %d: %v", user.UserID, err)
	}

	safeData := domain.SafeStaffData{
		UserID:    user.UserID,
		Username:  user.Username,
		Name:      user.Name,
		Role:      user.Role,
		Teaching:  subjects,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: user.DeletedAt,
	}

	return &safeData, nil
}

func (ur *userRepository) CreateSubject(ctx context.Context, subject *domain.Subject) error {
	var subjectVar domain.Subject
	err := ur.db.WithContext(ctx).Where("name = ?", subject.Name).First(&subjectVar).Error
	if err == nil {
		return fmt.Errorf("subject with %s name already exists", subject.Name)
	}

	var countSubjectCode int64
	err = ur.db.WithContext(ctx).Model(&domain.Subject{}).Where("subject_code = ?", subject.SubjectCode).Count(&countSubjectCode).Error
	if err != nil {
		return fmt.Errorf("could not check subject code: %v", err)
	}

	if countSubjectCode > 0 {
		return fmt.Errorf("subject code %s already exists", subject.SubjectCode)
	}

	err = ur.db.WithContext(ctx).Create(subject).Error
	if err != nil {
		return fmt.Errorf("could not create subject: %v", err)
	}

	return nil
}

func (ur *userRepository) GetSubjectDetail(ctx context.Context, id int) (*domain.Subject, error) {
	var subject domain.Subject
	err := ur.db.WithContext(ctx).Where("subject_id = ?", id).First(&subject).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subject not found")
		}
		return nil, fmt.Errorf("could not get subject details: %v", err)
	}

	return &subject, nil
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

func (ur *userRepository) GetAllSubject(ctx context.Context, userID int) (*[]domain.Subject, error) {
	var existingUser domain.User
	err := ur.db.WithContext(ctx).Where("user_id = ?", userID).First(&existingUser).Error
	if err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	var subjects []domain.Subject

	if existingUser.Role == "admin" {
		// Admin can see all subjects
		err = ur.db.WithContext(ctx).Find(&subjects).Error
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve all subjects: %w", err)
		}
	} else {
		// Non-admin users see only their assigned subjects
		err = ur.db.WithContext(ctx).
			Model(&existingUser).
			Association("Teaching").
			Find(&subjects)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve assigned subjects: %w", err)
		}
	}

	return &subjects, nil
}

func (ur *userRepository) UpdateSubject(ctx context.Context, id int, newSubjectData *domain.Subject) error {
	newSubjectData.UpdatedAt = time.Now()

	var countSubjectCode int64
	err := ur.db.WithContext(ctx).Model(&domain.Subject{}).Where("subject_code = ? AND subject_id != ?", newSubjectData.SubjectCode, id).Count(&countSubjectCode).Error
	if err != nil {
		return fmt.Errorf("could not check subject code: %v", err)
	}

	if countSubjectCode > 0 {
		return fmt.Errorf("subject with code %s already exists", newSubjectData.SubjectCode)
	}

	var countSubjectName int64
	err = ur.db.WithContext(ctx).Model(&domain.Subject{}).Where("name = ? AND subject_id != ?", newSubjectData.Name, id).Count(&countSubjectName).Error
	if err != nil {
		return fmt.Errorf("could not check subject name: %v", err)
	}

	if countSubjectName > 0 {
		return fmt.Errorf("subject with name %s already exists", newSubjectData.Name)
	}

	err = ur.db.WithContext(ctx).Model(&domain.Subject{}).
		Where("subject_id = ?", id).
		Updates(&newSubjectData).Error
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("subject with name %s already exists", newSubjectData.Name)
		}
		return fmt.Errorf("could not update staff: %v", err)
	}

	return nil

}

func (ur *userRepository) DeleteSubject(ctx context.Context, id int) error {
	// var subject domain.Subject
	// err := ur.db.WithContext(ctx).Where("subject_id = ? AND deleted_at IS NULL", id).First(&subject).Error
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return fmt.Errorf("subject not found")
	// 	}
	// 	return fmt.Errorf("could not get subject details: %v", err)
	// }

	// now := time.Now()
	// subject.DeletedAt = &now
	// err = ur.db.WithContext(ctx).Save(&subject).Error
	// if err != nil {
	// 	return fmt.Errorf("could not delete subject: %v", err)
	// }

	return nil
}

func (ur *userRepository) DeleteSubjectMass(ctx context.Context, ids *[]int) error {
	// var subjects []domain.Subject
	// err := ur.db.WithContext(ctx).
	// 	Where("subject_id IN (?) AND deleted_at IS NULL", *ids).
	// 	Find(&subjects).Error
	// if err != nil {
	// 	return fmt.Errorf("could not retrieve subject details: %v", err)
	// }

	// if len(subjects) == 0 {
	// 	return fmt.Errorf("no subject eligible for deletion")
	// }

	// now := time.Now()
	// err = ur.db.WithContext(ctx).
	// 	Model(&domain.Subject{}).
	// 	Where("subject_id IN (?)", *ids).
	// 	Update("deleted_at", now).Error
	// if err != nil {
	// 	return fmt.Errorf("could not delete subjects: %v", err)
	// }

	return nil
}

func (spr *userRepository) DeleteStaffMass(ctx context.Context, ids *[]int) error {
	var users []domain.User
	err := spr.db.WithContext(ctx).
		Where("user_id IN (?) AND deleted_at IS NULL", *ids).
		Find(&users).Error
	if err != nil {
		return fmt.Errorf("could not retrieve staff details: %v", err)
	}

	var staffToDelete []domain.User
	for _, user := range users {
		if user.Role != "admin" {
			staffToDelete = append(staffToDelete, user)
		}
	}

	if len(staffToDelete) == 0 {
		return fmt.Errorf("no staff eligible for deletion")
	}

	now := time.Now()
	err = spr.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("user_id IN ?", getIdsFromUsers(staffToDelete)).
		Update("deleted_at", now).Error
	if err != nil {
		return fmt.Errorf("could not delete staff: %v", err)
	}

	return nil
}

// Helper function to extract IDs from the filtered list of users
func getIdsFromUsers(users []domain.User) []int {
	ids := make([]int, len(users))
	for i, user := range users {
		ids[i] = user.UserID
	}
	return ids
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
