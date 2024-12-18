package repository

import (
	"context"
	"errors"
	"fmt"
	"notification/config"
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

func (spr *studentParentRepository) ApproveDCR(ctx context.Context, req map[string]interface{}) (*string, error) {
	fmt.Println("req payload")
	config.PrintStruct(req)
	now := time.Now()
	var parentHolder domain.Parent
	var existingParent domain.Parent
	var allocatedDataMsgs string

	// Start a database transaction
	tx := spr.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validate oldTelephone existence
	if req["oldTelephone"] == nil || req["oldTelephone"] == "" {
		tx.Rollback()
		return nil, fmt.Errorf("oldTelephone is required")
	}

	// Find the parent with oldTelephone
	err := tx.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", req["oldTelephone"]).First(&parentHolder).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, fmt.Errorf("parent not found with old telephone")
	} else if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error fetching parent: %v", err)
	}
	fmt.Println("parent found")
	config.PrintStruct(parentHolder)

	updateFields := make(map[string]interface{})

	// Update Name
	if name, ok := req["name"].(string); ok && strings.TrimSpace(name) != "" && name != parentHolder.Name {
		updateFields["name"] = strings.TrimSpace(name)
	}

	// Update Gender
	if gender, ok := req["gender"].(string); ok && strings.TrimSpace(gender) != "" && gender != parentHolder.Gender {
		updateFields["gender"] = strings.TrimSpace(gender)
	}

	// Update Telephone
	if telephone, ok := req["telephone"].(string); ok {
		trimmedTelephone := strings.TrimSpace(telephone)
		if trimmedTelephone != "" && trimmedTelephone != parentHolder.Telephone {
			updateFields["telephone"] = trimmedTelephone
			fmt.Println("masux - Telephone Updated:", trimmedTelephone)
		}
	}

	// Update Email
	if email, ok := req["email"].(string); ok && strings.TrimSpace(email) != "" {
		emailTrimmed := strings.TrimSpace(email)
		if parentHolder.Email == nil || *parentHolder.Email != emailTrimmed {
			updateFields["email"] = emailTrimmed
		}
	}

	// Always Update UpdatedAt
	updateFields["updated_at"] = now

	// Debug the updates to verify
	fmt.Println("Fields to Update:")
	config.PrintStruct(updateFields)

	// Find students associated with the parent
	studentsAssociated := []domain.Student{}
	err = tx.WithContext(ctx).Where("parent_id = ? AND deleted_at IS NULL", parentHolder.ParentID).Find(&studentsAssociated).Error
	if err != nil || len(studentsAssociated) == 0 {
		tx.Rollback()
		return nil, fmt.Errorf("no students associated with parent")
	}

	// Handle parent conflicts
	existingParent = domain.Parent{}
	err = tx.WithContext(ctx).Where("(name = ? OR telephone = ? OR email = ?) AND deleted_at IS NULL",
		req["name"], req["telephone"], req["email"]).First(&existingParent).Error

	if err == nil {
		// Conflict exists, update student associations
		var previousParentID int
		studentIDs := []int{}
		for _, student := range studentsAssociated {
			previousParentID = student.ParentID
			studentIDs = append(studentIDs, student.StudentID)
		}
		err = tx.WithContext(ctx).Model(&domain.Student{}).Where("student_id IN (?)", studentIDs).
			Updates(&domain.Student{ParentID: existingParent.ParentID}).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to reassign students: %v", err)
		}
		var parentCounter int64
		allocatedDataMsgs = fmt.Sprintf("Parent data already exists, allocating %d students to parent %s", len(studentsAssociated), existingParent.Name)
		err = tx.WithContext(ctx).Where("parent_id = ? AND deleted_at IS NULL", previousParentID).Count(&parentCounter).Error
		if err != nil {
			return nil, fmt.Errorf("error counting previous parent, error:%v", err)
		}
		if parentCounter > 0 {
			err = tx.WithContext(ctx).Where("parent_id = ? AND deleted_at IS NULL", previousParentID).Updates(&domain.Parent{
				DeletedAt: &now,
			}).Error
			if err != nil {
				return nil, fmt.Errorf("failed to delete student with no associated to any student, error:%v", err)
			}
		}
	}

	// Perform the update using GORM
	err = tx.Model(&domain.Parent{}).
		Where("parent_id = ? AND deleted_at IS NULL", parentHolder.ParentID).
		Updates(&updateFields).Error

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update parent: %v", err)
	}

	err = tx.Model(&domain.DataChangeRequest{}).Where("old_parent_telephone = ? AND is_reviewed IS false", req["oldTelephone"]).Updates(&domain.DataChangeRequest{
		IsReviewed: true,
	}).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update data change request: %v", err)
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	if allocatedDataMsgs != "" {
		fmt.Println("lol 0")

		return &allocatedDataMsgs, nil
	}

	fmt.Println("LOl")

	return nil, nil
}

func (spr *studentParentRepository) CreateStudentAndParent(ctx context.Context, req *domain.StudentAndParent) *[]string {
	var errList []string

	// Normalize email if provided
	if req.Parent.Email != nil {
		emailLowered := strings.ToLower(strings.TrimSpace(*req.Parent.Email))
		req.Parent.Email = &emailLowered
	}

	// Validate GradeLabel
	match, _ := regexp.MatchString("^[A-Za-z]+$", req.Student.GradeLabel)
	if !match {
		errList = append(errList, fmt.Sprintf("Invalid Grade Label: %s. Only letters (A-Z, a-z) are allowed.", req.Student.GradeLabel))
	}

	// Check for duplicate student telephone
	var studentCount int64
	err := spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ? AND deleted_at IS NULL", req.Student.Telephone).Count(&studentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student telephone: %v", err))
	} else if studentCount > 0 {
		errList = append(errList, fmt.Sprintf("Student with telephone %s already exists", req.Student.Telephone))
	}

	// Check for duplicate student name
	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("name = ? AND deleted_at IS NULL", req.Student.Name).Count(&studentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student name: %v", err))
	} else if studentCount > 0 {
		errList = append(errList, fmt.Sprintf("Student with name %s already exists", req.Student.Name))
	}

	// If student errors exist, return immediately
	if len(errList) > 0 {
		return &errList
	}

	// Find existing parent with matching attributes
	var existingParent domain.Parent
	err = spr.db.WithContext(ctx).Where(
		"(name = ? OR telephone = ? OR email = ?) AND deleted_at IS NULL",
		req.Parent.Name, req.Parent.Telephone, func() string {
			if req.Parent.Email != nil {
				return *req.Parent.Email
			}
			return ""
		}(),
	).First(&existingParent).Error

	// Check if parent condition matches
	oneToManyCondition := err == nil

	// Start transaction
	tx := spr.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return &[]string{fmt.Sprintf("Could not begin transaction: %v", tx.Error)}
	}

	if oneToManyCondition {
		// Parent exists, create student with reference
		req.Student.ParentID = existingParent.ParentID
		req.Student.CreatedAt = time.Now()
		req.Student.UpdatedAt = req.Student.CreatedAt
		req.Student.GradeLabel = strings.ToUpper(req.Student.GradeLabel)

		if err := tx.WithContext(ctx).Create(&req.Student).Error; err != nil {
			tx.Rollback()
			return &[]string{fmt.Sprintf("Could not insert student: %v", err)}
		}
	} else {
		// Create new parent
		req.Parent.CreatedAt = time.Now()
		req.Parent.UpdatedAt = req.Parent.CreatedAt

		if err := tx.WithContext(ctx).Create(&req.Parent).Error; err != nil {
			tx.Rollback()
			return &[]string{fmt.Sprintf("Could not insert parent: %v", err)}
		}

		// Create new student
		req.Student.ParentID = req.Parent.ParentID
		req.Student.CreatedAt = time.Now()
		req.Student.UpdatedAt = req.Student.CreatedAt
		req.Student.GradeLabel = strings.ToUpper(req.Student.GradeLabel)

		if err := tx.WithContext(ctx).Create(&req.Student).Error; err != nil {
			tx.Rollback()
			return &[]string{fmt.Sprintf("Could not insert student: %v", err)}
		}
	}

	// Commit transaction
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
		var studentExists domain.Student
		isDuplicate := false

		// Validate Parent Telephone
		if len(record.Parent.Telephone) > 15 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent telephone %s exceeds max length (15)", index+2, record.Parent.Telephone))
			isDuplicate = true
		}

		// Validate Student Telephone
		if len(record.Student.Telephone) > 15 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student telephone %s exceeds max length (15)", index+2, record.Student.Telephone))
			isDuplicate = true
		} else if err := spr.db.WithContext(ctx).Where("telephone = ? AND deleted_at IS NULL", record.Student.Telephone).First(&studentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student telephone %s already exists", index+2, record.Student.Telephone))
			isDuplicate = true
		}

		// Validate Student Name
		if err := spr.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", record.Student.Name).First(&studentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student name %s already exists", index+2, record.Student.Name))
			isDuplicate = true
		}

		// Skip records with validation errors
		if isDuplicate {
			continue
		}

		record.Parent.CreatedAt = now
		record.Parent.UpdatedAt = now
		record.Student.CreatedAt = now
		record.Student.UpdatedAt = now

		// Add valid records to readyToExecute
		readyToExecute = append(readyToExecute, record)
	}

	// If there are duplicate messages, return them
	if len(duplicateMessages) > 0 {
		return &duplicateMessages, nil
	}

	// Insert valid records into the database
	err := spr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range readyToExecute {
			var parentExist domain.Parent

			// Validate and handle parent
			err := tx.Where("(name = ? OR telephone = ? OR email = ?) AND deleted_at IS NULL",
				record.Parent.Name,
				record.Parent.Telephone,
				func() string {
					if record.Parent.Email != nil {
						return *record.Parent.Email
					}
					return ""
				}(),
			).First(&parentExist).Error

			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to query parent: %w", err)
			}

			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Parent does not exist, create new
				if err = tx.Create(&record.Parent).Error; err != nil {
					return fmt.Errorf("failed to insert parent: %w", err)
				}
			}

			// Use the existing or newly created parent ID for the student
			record.Student.ParentID = parentExist.ParentID
			if parentExist.ParentID == 0 {
				record.Student.ParentID = record.Parent.ParentID
			}

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
	var errList []string

	// Validate GradeLabel to only contain letters
	match, _ := regexp.MatchString("^[A-Za-z]+$", req.Student.GradeLabel)
	if !match {
		errList = append(errList, fmt.Sprintf("Invalid Grade Label: %s. Only letters (A-Z, a-z) are allowed.", req.Student.GradeLabel))
	}
	req.Student.GradeLabel = strings.ToUpper(req.Student.GradeLabel)

	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		errList = append(errList, fmt.Sprintf("could not begin transaction: %v", err))
		return &errList
	}

	now := time.Now()
	req.Student.UpdatedAt = now
	req.Parent.UpdatedAt = now

	// Check for duplicate student data
	if req.Student.Name != "" {
		err := spr.db.WithContext(ctx).
			Where("name = ? AND student_id != ? AND deleted_at IS NULL", req.Student.Name, id).
			First(&domain.Student{}).Error
		if err == nil {
			errList = append(errList, fmt.Sprintf("Student with name %s already exists", req.Student.Name))
		}
	}
	if req.Student.Telephone != "" {
		err := spr.db.WithContext(ctx).
			Where("telephone = ? AND student_id != ? AND deleted_at IS NULL", req.Student.Telephone, id).
			First(&domain.Student{}).Error
		if err == nil {
			errList = append(errList, fmt.Sprintf("Student with telephone %s already exists", req.Student.Telephone))
		}
	}

	if len(errList) > 0 {
		tx.Rollback()
		return &errList
	}

	// Fetch existing student with parent
	err := spr.db.WithContext(ctx).Preload("Parent").Where("student_id = ? AND deleted_at IS NULL", id).First(&student).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errList = append(errList, fmt.Sprintf("can't find student with id %d", id))
			return &errList
		}
		errList = append(errList, fmt.Sprintf("database error: %v", err))
		return &errList
	}

	// Build map of updated fields for student
	updatedStudentFields := make(map[string]interface{})
	if req.Student.Name != "" && req.Student.Name != student.Name {
		updatedStudentFields["name"] = req.Student.Name
	}
	if req.Student.Grade != 0 && req.Student.Grade != student.Grade {
		updatedStudentFields["grade"] = req.Student.Grade
	}
	if req.Student.GradeLabel != "" && req.Student.GradeLabel != student.GradeLabel {
		updatedStudentFields["grade_label"] = req.Student.GradeLabel
	}
	if req.Student.Gender != "" && req.Student.Gender != student.Gender {
		updatedStudentFields["gender"] = req.Student.Gender
	}
	if req.Student.Telephone != "" && req.Student.Telephone != student.Telephone {
		updatedStudentFields["telephone"] = req.Student.Telephone
	}
	if len(updatedStudentFields) > 0 {
		updatedStudentFields["updated_at"] = now
	}

	// Build map of updated fields for parent
	updatedParentFields := make(map[string]interface{})

	if req.Parent.Name != "" && req.Parent.Name != student.Parent.Name {
		updatedParentFields["name"] = req.Parent.Name
	}
	if req.Parent.Gender != "" && req.Parent.Gender != student.Parent.Gender {
		updatedParentFields["gender"] = req.Parent.Gender
	}
	if req.Parent.Telephone != "" && req.Parent.Telephone != student.Parent.Telephone {
		updatedParentFields["telephone"] = req.Parent.Telephone
	}
	if req.Parent.Email != nil && (student.Parent.Email == nil || *req.Parent.Email != *student.Parent.Email) {
		updatedParentFields["email"] = *req.Parent.Email
	}
	if len(updatedParentFields) > 0 {
		updatedParentFields["updated_at"] = now
	}

	if len(updatedParentFields) > 0 {
		var existingParent domain.Parent
		err = tx.WithContext(ctx).Where(
			"(name = ? OR telephone = ? OR email = ?) AND deleted_at IS NULL",
			updatedParentFields["name"], updatedParentFields["telephone"], updatedParentFields["email"],
		).First(&existingParent).Error

		if err == nil {
			// Update parent_id to the existing parent's ID
			updatedStudentFields["ParentID"] = existingParent.ParentID
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Update the parent and keep the current parent_id
			if err := tx.WithContext(ctx).
				Model(domain.Parent{}).
				Where("parent_id = ? AND deleted_at IS NULL", student.ParentID).
				Updates(updatedParentFields).Error; err != nil {
				tx.Rollback()
				errList = append(errList, fmt.Sprintf("failed to update parent: %v", err))
				return &errList
			}
		} else {
			tx.Rollback()
			errList = append(errList, fmt.Sprintf("database error while checking parent: %v", err))
			return &errList
		}
	}

	config.PrintStruct(updatedStudentFields)
	err = tx.WithContext(ctx).
		Model(domain.Student{}).
		Where("student_id = ? AND deleted_at IS NULL", student.StudentID).
		Updates(updatedStudentFields).Error
	if err != nil {
		tx.Rollback()
		errList = append(errList, fmt.Sprintf("failed to update student: %v", err))
		return &errList
	}

	if err := tx.Commit().Error; err != nil {
		errList = append(errList, fmt.Sprintf("could not commit transaction: %v", err))
		return &errList
	}

	return nil
}

func (spr *studentParentRepository) SPMassDelete(ctx context.Context, studentIDs *[]int) error {
	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	currentTime := time.Now()

	// Iterate over the student IDs to process each student
	for _, studentID := range *studentIDs {
		// Retrieve the parent_id from the student record
		var student domain.Student
		err := tx.WithContext(ctx).
			Where("student_id = ? AND deleted_at IS NULL", studentID).
			First(&student).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("student with ID %d not found", studentID)
			}
			return fmt.Errorf("error retrieving student: %v", err)
		}

		// Count remaining active students associated with the same parent
		var remainingStudentCount int64
		err = tx.WithContext(ctx).
			Model(&domain.Student{}).
			Where("parent_id = ? AND deleted_at IS NULL", student.ParentID).
			Count(&remainingStudentCount).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error counting remaining students: %v", err)
		}

		if remainingStudentCount > 1 {
			// Soft delete only the student by setting DeletedAt
			err = tx.WithContext(ctx).
				Model(&domain.Student{}).
				Where("student_id = ?", studentID).
				Update("deleted_at", currentTime).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting student: %v", err)
			}
		} else {
			// Soft delete both the student and the parent by setting DeletedAt
			err = tx.WithContext(ctx).
				Model(&domain.Student{}).
				Where("student_id = ?", studentID).
				Update("deleted_at", currentTime).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting student: %v", err)
			}

			err = tx.WithContext(ctx).
				Model(&domain.Parent{}).
				Where("parent_id = ?", student.ParentID).
				Update("deleted_at", currentTime).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting parent: %v", err)
			}
		}
	}

	// Commit the transaction
	err := tx.Commit().Error
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

	// Retrieve the parent_id from the student record
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

	// Count students associated with the same parent
	var studentCount int64
	err = tx.WithContext(ctx).
		Model(&domain.Student{}).
		Where("parent_id = ? AND deleted_at IS NULL", student.ParentID).
		Count(&studentCount).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error counting students: %v", err)
	}

	currentTime := time.Now()

	if studentCount > 1 {
		// Soft delete only the student by setting DeletedAt
		err = tx.WithContext(ctx).
			Model(&domain.Student{}).
			Where("student_id = ?", studentID).
			Update("deleted_at", currentTime).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error soft deleting student: %v", err)
		}
	} else {
		// Soft delete both the student and the parent by setting DeletedAt
		err = tx.WithContext(ctx).
			Model(&domain.Student{}).
			Where("student_id = ?", studentID).
			Update("deleted_at", currentTime).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error soft deleting student: %v", err)
		}

		err = tx.WithContext(ctx).
			Model(&domain.Parent{}).
			Where("parent_id = ?", student.ParentID).
			Update("deleted_at", currentTime).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error soft deleting parent: %v", err)
		}
	}

	// Commit the transaction
	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (spr *studentParentRepository) DeleteStudentAndParentMass(ctx context.Context, studentIDs *[]int) error {
	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}

	currentTime := time.Now()

	// Iterate over the student IDs to process each student
	for _, studentID := range *studentIDs {
		// Retrieve the parent_id from the student record
		var student domain.Student
		err := tx.WithContext(ctx).
			Where("student_id = ? AND deleted_at IS NULL", studentID).
			First(&student).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("student with ID %d not found", studentID)
			}
			return fmt.Errorf("error retrieving student: %v", err)
		}

		// Count remaining active students associated with the same parent
		var remainingStudentCount int64
		err = tx.WithContext(ctx).
			Model(&domain.Student{}).
			Where("parent_id = ? AND deleted_at IS NULL", student.ParentID).
			Count(&remainingStudentCount).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error counting remaining students: %v", err)
		}

		if remainingStudentCount > 1 {
			// Soft delete only the student by setting DeletedAt
			err = tx.WithContext(ctx).
				Model(&domain.Student{}).
				Where("student_id = ?", studentID).
				Update("deleted_at", currentTime).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting student: %v", err)
			}
		} else {
			// Soft delete both the student and the parent by setting DeletedAt
			err = tx.WithContext(ctx).
				Model(&domain.Student{}).
				Where("student_id = ?", studentID).
				Update("deleted_at", currentTime).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting student: %v", err)
			}

			err = tx.WithContext(ctx).
				Model(&domain.Parent{}).
				Where("parent_id = ?", student.ParentID).
				Update("deleted_at", currentTime).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error soft deleting parent: %v", err)
			}
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
		Preload("Parent", func(db *gorm.DB) *gorm.DB {
			// Apply a filter to exclude parents with deleted_at NOT NULL
			return db.WithContext(ctx).Where("deleted_at IS NULL")
		}).
		Where("students.student_id = ? AND students.deleted_at IS NULL", studentID).
		First(&result.Student).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("student with ID %d not found", studentID)
		}
		return nil, fmt.Errorf("could not fetch student details: %v", err)
	}

	// Explicitly check if the parent was not loaded
	if result.Student.ParentID != 0 && result.Student.Parent.DeletedAt != nil {
		result.Student.Parent = domain.Parent{} // Reset to an empty struct if deleted
	}

	return &result, nil
}

func (spr *studentParentRepository) GetAllDataChangeRequestByID(ctx context.Context, dcrID int) (*domain.DataChangeRequest, error) {
	var result domain.DataChangeRequest

	err := spr.db.WithContext(ctx).
		Where("request_id = ? AND is_reviewed IS FALSE", dcrID).
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("data change request with ID %d not found", dcrID)
		}
		return nil, fmt.Errorf("could not fetch data change request details: %v", err)
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

func (spr *studentParentRepository) ReviewDCR(ctx context.Context, dcrID int) error {
	result := spr.db.WithContext(ctx).
		Model(&domain.DataChangeRequest{}).
		Where("request_id = ?", dcrID).
		Update("is_reviewed", true)

	if result.Error != nil {
		return fmt.Errorf("failed to update is_reviewed for request_id %d: %w", dcrID, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no data change request found with request_id %d", dcrID)
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
