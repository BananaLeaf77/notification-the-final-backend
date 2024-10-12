package repository

import (
	"context"
	"errors"
	"fmt"
	"notification/domain"
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
	// Check if the student telephone already exists
	var errList []string
	var existingStudent domain.Student
	err := spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", req.Student.Telephone).First(&existingStudent).Error
	// jika query diatas berhasil berarti error nya nil!!
	if err == nil {
		errList = append(errList, fmt.Sprintf("student with telephone %s already exists", req.Student.Telephone))

	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errList = append(errList, fmt.Sprintf("error checking student telephone: %v", err))
	}

	// Check if the parent telephone already exists
	var existingParent domain.Parent
	err = spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", req.Parent.Telephone).First(&existingParent).Error
	if err == nil {
		errList = append(errList, fmt.Sprintf("parent with telephone %s already exists", req.Parent.Telephone))

	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errList = append(errList, fmt.Sprintf("error checking parent telephone: %v", err))

	}

	// Check if the parent email already exists
	err = spr.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", req.Parent.Email).First(&existingParent).Error
	if err == nil {
		errList = append(errList, fmt.Sprintf("parent with email %s already exists", *req.Parent.Email))

	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errList = append(errList, fmt.Sprintf("error checking parent email: %v", err))

	}

	// Check if the student name already exists
	err = spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", req.Student.Name).First(&existingStudent).Error
	if err == nil {
		errList = append(errList, fmt.Sprintf("student with name %s already exists", req.Student.Name))

	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errList = append(errList, fmt.Sprintf("error checking student name: %v", err))

	}

	// Check if the parent name already exists
	err = spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", req.Parent.Name).First(&existingParent).Error
	if err == nil {
		errList = append(errList, fmt.Sprintf("parent with name %s already exists", req.Parent.Name))

	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		errList = append(errList, fmt.Sprintf("error checking parent name: %v", err))
	}

	if len(errList) > 0 {
		return &errList
	}

	// If all checks pass, proceed with the transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		var singleArr []string
		singleArr = append(singleArr, fmt.Sprintf("could not begin transaction: %v", err))
		return &singleArr
	}

	// Insert parent
	req.Parent.CreatedAt = time.Now()
	req.Parent.UpdatedAt = req.Parent.CreatedAt
	if err = tx.WithContext(ctx).Create(&req.Parent).Error; err != nil {
		tx.Rollback()
		var singleArr []string
		singleArr = append(singleArr, fmt.Sprintf("could not insert parent: %v", err))
		return &singleArr
	}

	// Set the ParentID after inserting the parent
	req.Student.ParentID = req.Parent.ParentID

	// Insert student
	req.Student.CreatedAt = time.Now()
	req.Student.UpdatedAt = req.Student.CreatedAt
	if err = tx.WithContext(ctx).Create(&req.Student).Error; err != nil {
		tx.Rollback()
		var singleArr []string
		singleArr = append(singleArr, fmt.Sprintf("could not insert student: %v", err))
		return &singleArr
	}

	// Commit the transaction
	if err = tx.Commit().Error; err != nil {
		var singleArr []string
		singleArr = append(singleArr, fmt.Sprintf("could not commit transaction: %v", err))
		return &singleArr
	}

	return nil
}

func (spr *studentParentRepository) ImportCSV(ctx context.Context, payload *[]domain.StudentAndParent) (*[]string, error) {
	var duplicateMessages []string

	now := time.Now()

	for index, record := range *payload {
		// Check if parent already exists by telephone
		var parentExists domain.Parent
		err := spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", record.Parent.Telephone).First(&parentExists).Error
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with telephone %s already exists", index+1, record.Parent.Telephone))
			continue
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("row %d: error checking if parent exists by telephone: %v", index+1, err)
		}

		// Check if parent already exists by email
		err = spr.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", record.Parent.Email).First(&parentExists).Error
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with email %s already exists", index+1, *record.Parent.Email))
			continue
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("row %d: error checking if parent exists by email: %v", index+1, err)
		}

		// Check if parent already exists by name
		err = spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", record.Parent.Name).First(&parentExists).Error
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent with name %s already exists", index+1, record.Parent.Name))
			continue
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("row %d: error checking if parent exists by name: %v", index+1, err)
		}

		// Check if student already exists by telephone
		var studentExists domain.Student
		err = spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", record.Student.Telephone).First(&studentExists).Error
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with telephone %s already exists", index+1, record.Student.Telephone))
			continue
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("row %d: error checking if student exists: %v", index+1, err)
		}

		// Check if student already exists by name
		err = spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", record.Student.Name).First(&studentExists).Error
		if err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student with name %s already exists", index+1, record.Student.Name))
			continue
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("row %d: error checking if student exists by name: %v", index+1, err)
		}

		// Insert parent and student
		tx := spr.db.Begin()
		if err := tx.Error; err != nil {
			return nil, fmt.Errorf("could not begin transaction: %v", err)
		}

		record.Parent.CreatedAt = now
		record.Parent.UpdatedAt = now
		err = tx.WithContext(ctx).Create(&record.Parent).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("row %d: could not insert parent: %v", index+1, err)
		}

		record.Student.ParentID = record.Parent.ParentID
		record.Student.CreatedAt = now
		record.Student.UpdatedAt = now
		err = tx.WithContext(ctx).Create(&record.Student).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("row %d: could not insert student: %v", index+1, err)
		}

		tx.Commit()
	}

	if len(duplicateMessages) > 0 {
		return &duplicateMessages, nil
	}

	return nil, nil
}

func (spr *studentParentRepository) UpdateStudentAndParent(ctx context.Context, req *domain.StudentAndParent) error {
	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	// Check if student exists
	var student domain.Student
	err := tx.WithContext(ctx).Where("id = ?", req.Student.StudentID).First(&student).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("student with ID %d not found", req.Student.StudentID)
		}
		return fmt.Errorf("error finding student: %v", err)
	}

	// Check if parent exists
	var parent domain.Parent
	err = tx.WithContext(ctx).Where("id = ?", req.Parent.ParentID).First(&parent).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("parent with ID %d not found", req.Parent.ParentID)
		}
		return fmt.Errorf("error finding parent: %v", err)
	}

	// Update parent details
	req.Parent.UpdatedAt = time.Now()
	err = tx.WithContext(ctx).Model(&parent).Updates(req.Parent).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not update parent: %v", err)
	}

	// Update student details
	req.Student.UpdatedAt = time.Now()
	err = tx.WithContext(ctx).Model(&student).Updates(req.Student).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not update student: %v", err)
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
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
	err = tx.WithContext(ctx).Where("id = ?", student.ParentID).Delete(&domain.Parent{}).Error
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

func (spr *studentParentRepository) GetStudentDetailsByID(ctx context.Context, studentID int) (*domain.StudentAndParent, error) {
	var result domain.StudentAndParent

	// Use Preload to load the Parent data automatically and use the correct column student_id
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
