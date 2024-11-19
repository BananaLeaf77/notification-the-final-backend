package repository

import (
	"context"
	"errors"
	"fmt"
	"notification/domain"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type studentParentRepository struct {
	db *gorm.DB
}

func NewStudentParentRepository(database *gorm.DB) domain.StudentParentRepo {
	return &studentParentRepository{
		db: database,
	}
}

func (spr *studentParentRepository) CreateStudentAndParent(ctx context.Context, req *domain.StudentAndParent) *[]string {
	var errList []string

	if req.Parent.Email != nil {
		emailLowered := strings.ToLower(strings.TrimSpace(*req.Parent.Email))
		req.Parent.Email = &emailLowered
	}

	match, _ := regexp.MatchString("^[A-Za-z]+$", req.Student.GradeLabel)
	if !match {
		errList = append(errList, fmt.Sprintf("Invalid Grade Label: %s. Only letters (A-Z, a-z) are allowed.", req.Student.GradeLabel))
	}

	var studentCount int64
	err := spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ? AND deleted_at IS NULL", req.Student.Telephone).Count(&studentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student telephone: %v", err))
	} else if studentCount > 0 {
		errList = append(errList, fmt.Sprintf("Student with telephone %s already exists", req.Student.Telephone))
	}

	var parentCount int64
	err = spr.db.WithContext(ctx).Model(&domain.Parent{}).Where("telephone = ? AND deleted_at IS NULL", req.Parent.Telephone).Count(&parentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for parent telephone: %v", err))
	} else if parentCount > 0 {
		errList = append(errList, fmt.Sprintf("Parent with telephone %s already exists", req.Parent.Telephone))
	}

	if req.Parent.Email != nil && *req.Parent.Email != "" {
		err = spr.db.WithContext(ctx).Model(&domain.Parent{}).Where("email = ? AND deleted_at IS NULL", *req.Parent.Email).Count(&parentCount).Error
		if err != nil {
			errList = append(errList, fmt.Sprintf("Error checking for parent email: %v", err))
		} else if parentCount > 0 {
			errList = append(errList, fmt.Sprintf("Parent with email %s already exists", *req.Parent.Email))
		}
	}

	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("name = ? AND deleted_at IS NULL", req.Student.Name).Count(&studentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student name: %v", err))
	} else if studentCount > 0 {
		errList = append(errList, fmt.Sprintf("Student with name %s already exists", req.Student.Name))
	}

	err = spr.db.WithContext(ctx).Model(&domain.Parent{}).Where("name = ? AND deleted_at IS NULL", req.Parent.Name).Count(&parentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for parent name: %v", err))
	} else if parentCount > 0 {
		errList = append(errList, fmt.Sprintf("Parent with name %s already exists", req.Parent.Name))
	}

	if len(errList) > 0 {
		return &errList
	}

	tx := spr.db.Begin()
	if tx.Error != nil {
		return &[]string{fmt.Sprintf("Could not begin transaction: %v", tx.Error)}
	}

	// Insert Parent
	req.Parent.CreatedAt = time.Now()
	req.Parent.UpdatedAt = req.Parent.CreatedAt
	if err := tx.WithContext(ctx).Create(&req.Parent).Error; err != nil {
		tx.Rollback()
		return &[]string{fmt.Sprintf("Could not insert parent: %v", err)}
	}

	// Set ParentID for Student
	req.Student.ParentID = req.Parent.ParentID
	req.Student.CreatedAt = time.Now()
	req.Student.UpdatedAt = req.Student.CreatedAt

	// Insert Student
	if err := tx.WithContext(ctx).Create(&req.Student).Error; err != nil {
		tx.Rollback()
		return &[]string{fmt.Sprintf("Could not insert student: %v", err)}
	}

	// Commit Transaction
	if err := tx.Commit().Error; err != nil {
		return &[]string{fmt.Sprintf("Could not commit transaction: %v", err)}
	}

	return nil
}

func (spr *studentParentRepository) ImportCSV(ctx context.Context, payload *[]domain.StudentAndParent) (*[]string, error) {
	var duplicateMessages []string
	var readyToExecute []domain.StudentAndParent
	now := time.Now()

	// Validate and filter the records
	for index, record := range *payload {
		var parentExists domain.Parent
		var studentExists domain.Student

		// Check for errors independently for each field
		isDuplicate := false

		// Check Parent Telephone
		if len(record.Parent.Telephone) > 15 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with telephone %s is too long", index+2, record.Parent.Telephone))
			isDuplicate = true
		} else if err := spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", record.Parent.Telephone).First(&parentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with telephone %s already exists", index+2, record.Parent.Telephone))
			isDuplicate = true
		}

		// Check Parent Name
		if err := spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", record.Parent.Name).First(&parentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with name %s already exists", index+2, record.Parent.Name))
			isDuplicate = true
		}

		// Check Parent Email
		if record.Parent.Email != nil && *record.Parent.Email != "" {
			parentEmailLowered := strings.ToLower(*record.Parent.Email)
			if err := spr.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", parentEmailLowered).First(&parentExists).Error; err == nil {
				duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with email %s already exists", index+2, parentEmailLowered))
				isDuplicate = true
			}
			record.Parent.Email = &parentEmailLowered
		}

		// Check Student Telephone
		if len(record.Student.Telephone) > 15 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with telephone %s is too long", index+2, record.Student.Telephone))
			isDuplicate = true
		} else if err := spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", record.Student.Telephone).First(&studentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with telephone %s already exists", index+2, record.Student.Telephone))
			isDuplicate = true
		}

		// Check Student Name
		if err := spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", record.Student.Name).First(&studentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with name %s already exists", index+2, record.Student.Name))
			isDuplicate = true
		}

		// Skip the record if any duplication was found
		if isDuplicate {
			continue
		}
		record.Parent.CreatedAt = now
		record.Parent.UpdatedAt = now
		record.Student.CreatedAt = now
		record.Student.UpdatedAt = now
		// If no duplicates found, add to readyToExecute
		readyToExecute = append(readyToExecute, record)
	}

	// If there are duplicate messages, return them
	if len(duplicateMessages) > 0 {
		fmt.Println(duplicateMessages)
		return &duplicateMessages, nil
	}

	// Insert the valid records into the database
	err := spr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range readyToExecute {
			// Insert parent
			if err := tx.Create(&record.Parent).Error; err != nil {
				return fmt.Errorf("failed to insert parent: %w", err)
			}

			// Assign the parent ID to the student
			record.Student.ParentID = record.Parent.ParentID

			// Insert student
			if err := tx.Create(&record.Student).Error; err != nil {
				return fmt.Errorf("failed to insert student: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute database transaction: %w", err)
	}

	return nil, nil
}

func (spr *studentParentRepository) UpdateStudentAndParent(ctx context.Context, id int, req *domain.StudentAndParent) *[]string {
	var student domain.Student
	var duplicatedDataStudent domain.Student
	var duplicatedDataParent domain.Parent

	var errList []string

	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		errList = append(errList, fmt.Sprintf("could not begin transaction: %v", err))
		return &errList
	}

	now := time.Now()
	req.Student.UpdatedAt = now
	req.Student.Parent.UpdatedAt = now

	// Check if the student exists
	err := spr.db.WithContext(ctx).Where("student_id = ? AND deleted_at IS NULL", id).First(&student).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errList = append(errList, fmt.Sprintf("can't find student with id %d", id))
			return &errList
		} else {
			errList = append(errList, fmt.Sprintf("database error: %v", err))
			return &errList
		}
	}

	// Ensure no duplicate student fields
	checkUniqueStudentField := func(field, value string) {
		err := spr.db.WithContext(ctx).Where(fmt.Sprintf("%s = ? AND student_id != ? AND deleted_at IS NULL", field), value, id).First(&duplicatedDataStudent).Error
		if err == nil {
			errList = append(errList, fmt.Sprintf("Student with %s %s already exists", field, value))
		}
	}

	checkUniqueStudentField("name", req.Student.Name)
	checkUniqueStudentField("telephone", req.Student.Telephone)

	// Ensure no duplicate parent fields
	checkUniqueParentField := func(field, value string) {
		err := spr.db.WithContext(ctx).Where(fmt.Sprintf("%s = ? AND parent_id != ? AND deleted_at IS NULL", field), value, student.ParentID).First(&duplicatedDataParent).Error
		if err == nil {
			errList = append(errList, fmt.Sprintf("Parent with %s %s already exists", field, value))
		}
	}

	checkUniqueParentField("name", req.Parent.Name)

	if req.Parent.Email != nil && *req.Parent.Email != "" {
		loweredEmail := strings.ToLower(*req.Parent.Email)
		req.Parent.Email = &loweredEmail
		checkUniqueParentField("email", loweredEmail)
	}

	checkUniqueParentField("telephone", req.Parent.Telephone)

	if len(errList) > 0 {
		return &errList
	}

	// Update only if parent exists, or create otherwise
	if err := tx.WithContext(ctx).
		Where("parent_id = ?", student.ParentID).
		Assign(req.Parent).
		FirstOrCreate(&req.Student.Parent).
		Error; err != nil {
		tx.Rollback()
		errList = append(errList, fmt.Sprintf("could not update parent: %v", err))
		return &errList
	}

	// Update the student
	if err := tx.WithContext(ctx).
		Model(&req.Student).
		Where("student_id = ? AND parent_id = ?", id, student.ParentID).
		Updates(req.Student).
		Error; err != nil {
		tx.Rollback()
		errList = append(errList, fmt.Sprintf("could not update student: %v", err))
		return &errList
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		errList = append(errList, fmt.Sprintf("could not commit transaction: %v", err))
		return &errList
	}

	return nil
}

func (spr *studentParentRepository) DeleteStudentAndParent(ctx context.Context, studentID int) error {
	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	// Retrieve the parentID from the student record
	var student domain.Student
	err := tx.WithContext(ctx).
		Select("parent_id").
		Where("student_id = ? AND deleted_at IS NULL", studentID).
		First(&student).Error

	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("student with ID %d not found", studentID)
		}
		return fmt.Errorf("error retrieving student: %v", err)
	}

	// Soft delete the student
	err = tx.WithContext(ctx).Where("student_id = ?", studentID).Delete(&domain.Student{}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error deleting student: %v", err)
	}

	// Soft delete the parent
	err = tx.WithContext(ctx).Where("parent_id = ?", student.ParentID).Delete(&domain.Parent{}).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("parent with ID %d not found", student.ParentID)
		}
		return fmt.Errorf("error deleting parent: %v", err)
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (spr *studentParentRepository) DeleteStudentAndParentMass(ctx context.Context, studentIDS *[]int) error {
	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	// Iterate over the student IDs to delete them in bulk
	for _, studentID := range *studentIDS {
		var student domain.Student
		// Retrieve the parentID from the student record
		err := tx.WithContext(ctx).
			Select("parent_id").
			Where("student_id = ? AND deleted_at IS NULL", studentID).
			First(&student).Error

		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("student with ID %d not found", studentID)
			}
			return fmt.Errorf("error retrieving student: %v", err)
		}

		// Soft delete the student
		err = tx.WithContext(ctx).Where("student_id = ?", studentID).Delete(&domain.Student{}).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error deleting student with ID %d: %v", studentID, err)
		}

		// Soft delete the parent if no other students are linked to this parent
		err = tx.WithContext(ctx).
			Where("parent_id = ?", student.ParentID).
			Where("NOT EXISTS (SELECT 1 FROM students WHERE parent_id = ? AND deleted_at IS NULL)", student.ParentID).
			Delete(&domain.Parent{}).Error

		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("parent with ID %d not found", student.ParentID)
			}
			return fmt.Errorf("error deleting parent with ID %d: %v", student.ParentID, err)
		}
	}

	// Commit the transaction
	err := tx.Commit().Error
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (spr *studentParentRepository) GetStudentDetailsByID(ctx context.Context, studentID int) (*domain.StudentAndParent, error) {
	var result domain.StudentAndParent

	err := spr.db.WithContext(ctx).
		Preload("Parent").
		Where("students.student_id = ? AND students.deleted_at IS NULL", studentID).
		First(&result.Student).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("student with ID %d not found", studentID)
		}
		return nil, fmt.Errorf("could not fetch student details: %v", err)
	}

	return &result, nil
}

func (spr *studentParentRepository) DataChangeRequest(ctx context.Context, datas domain.DataChangeRequest) error {
	err := spr.db.WithContext(ctx).Create(&datas).Error
	if err != nil {
		return err
	}

	return nil
}

func (spr *studentParentRepository) GetAllDataChangeRequest(ctx context.Context) (*[]domain.DataChangeRequest, error) {
	var req []domain.DataChangeRequest

	if err := spr.db.WithContext(ctx).Where("is_reviewed IS FALSE").Find(&req).Error; err != nil {
		return nil, fmt.Errorf("could not get all data change request : %v", err)
	}

	return &req, nil
}

// func (spr *studentParentRepository) GetClassIDByName(className string) (*int, error) {
// 	var class domain.Class

// 	err := spr.db.Where("name = ?", className).First(&class).Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			return nil, fmt.Errorf("class not found: %s", className)
// 		}
// 		return nil, fmt.Errorf("error retrieving class: %v", err)
// 	}

// 	return &class.ClassID, nil
// }
