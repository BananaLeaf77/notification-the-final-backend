package repository

import (
	"context"
	"fmt"
	"notification/domain"
	"gorm.io/gorm"
	"os"
)

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository(database *gorm.DB) domain.StudentRepo {
	return &studentRepository{
		db: database,
	}
}

func (sp *studentRepository) GetAllStudent(ctx context.Context) (*[]domain.Student, error) {
	var students []domain.Student

	if err := sp.db.WithContext(ctx).Where("deleted_at IS NULL").Find(&students).Error; err != nil {
		return nil, fmt.Errorf("could not get all students: %v", err)
	}

	return &students, nil
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
