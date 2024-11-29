package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"os"

	"gorm.io/gorm"
)

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(database *gorm.DB) domain.StudentRepo {
	return &studentRepository{
		db: database,
	}
}

func (spr *studentRepository) GetStudentByParentTelephone(ctx context.Context, parTel string) (*domain.StudentAndParent, error) {
	// Define a result struct to hold the combined data
	var result domain.StudentAndParent
	var res struct {
		StudentID    int     `json:"student_id"`
		StudentName  string  `json:"student_name"`
		StudentGrade int     `json:"student_grade"`
		ParentID     int     `json:"parent_id"`
		ParentName   string  `json:"parent_name"`
		ParentTel    string  `json:"parent_telephone"`
		ParentEmail  *string `json:"parent_email"`
	}

	// Perform an explicit join query
	err := spr.db.WithContext(ctx).
		Table("students").
		Select("students.*, parents.*").
		Joins("JOIN parents ON parents.parent_id = students.parent_id").
		Where("parents.telephone = ? AND parents.deleted_at IS NULL", parTel).
		Scan(&res).Error

	// Handle errors
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no student found for parent telephone %s", parTel)
		}
		return nil, fmt.Errorf("could not fetch student details: %v", err)
	}

	err = spr.db.WithContext(ctx).Preload("Parent").Where("student_id = ? AND deleted_at IS NULL", res.StudentID).First(&result.Student).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (sp *studentRepository) GetAllStudent(ctx context.Context, userID int) (*[]domain.Student, error) {
	// Check if the user exists
	var existingUser domain.User
	err := sp.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Preload("Teaching").First(&existingUser).Error
	if err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	var students []domain.Student

	if existingUser.Role == "admin" {
		// Admin gets all students
		err = sp.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&students).Error
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve all students: %w", err)
		}
	} else {
		// Non-admin gets students in grades they are assigned to teach
		if len(existingUser.Teaching) == 0 {
			return &students, nil // Return empty list if no teaching subjects are found
		}

		// Extract grades from the Teaching association
		var grades []int
		for _, subject := range existingUser.Teaching {
			grades = append(grades, subject.Grade)
		}

		// Remove duplicates from grades
		grades = uniqueIntSlice(grades)

		// Fetch students whose grade matches the grades the user teaches
		err = sp.db.WithContext(ctx).
			Where("grade IN ? AND deleted_at IS NULL", grades).
			Find(&students).Error
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve students: %w", err)
		}
	}

	return &students, nil
}

// Helper function to remove duplicate integers from a slice
func uniqueIntSlice(input []int) []int {
	keys := make(map[int]bool)
	var list []int
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (sp *studentRepository) DownloadInputDataTemplate(ctx context.Context) (*string, error) {
	filePath := "./template/input_data_template.csv"

	// Check if the file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template file not found: %v", err)
		}
		return nil, err
	}

	return &filePath, nil
}
