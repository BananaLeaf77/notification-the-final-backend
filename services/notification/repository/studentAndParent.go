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

// func (spr *studentParentRepository) DeleteDCR(ctx context.Context, dcrID int) error{
// 	var dcr domain.DataChangeRequest
// 	err := spr.db.WithContext(ctx).Model(&domain.DataChangeRequest{}).Where("request_id = ?", dcrID).First(&dcr).Error
// 	if err != nil {
// 		return err
// 	}

// 	err = spr.db.WithContext(ctx).Model(&domain.DataChangeRequest{}).Where("request_id = ?", )

// 	return nil
// }

func (spr *studentParentRepository) ApproveDCR(ctx context.Context, req map[string]interface{}) (*string, error) {
	var dcr domain.DataChangeRequest
	var Parent domain.Parent
	var AssociatedStudent []domain.Student
	tNow := time.Now()
	var comparedData struct {
		Name      string
		Gender    string
		Telephone string
		Email     *string
		UpdatedAt time.Time
	}

	// Begin transaction
	tx := spr.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			fmt.Print("==========================================================")
			tx.Rollback()
			panic(r)
		}
	}()

	// Find the parent by oldTelephone
	oldTelephone := req["oldTelephone"]
	err := tx.Where("telephone = ?", oldTelephone).First(&Parent).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failure on finding parent with old telephone: %s, error: %v", oldTelephone, err)
	}

	// Find the DataChangeRequest for the given oldTelephone
	err = tx.Where("old_parent_telephone = ? AND is_reviewed IS FALSE", oldTelephone).First(&dcr).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failure finding data change request, error: %v", err)
	}

	// Compare dcr and parent fields to populate comparedData
	if dcr.NewParentName != nil && *dcr.NewParentName != Parent.Name {
		comparedData.Name = *dcr.NewParentName
	} else {
		comparedData.Name = Parent.Name
	}

	if dcr.NewParentGender != nil && *dcr.NewParentGender != Parent.Gender {
		comparedData.Gender = *dcr.NewParentGender
	} else {
		comparedData.Gender = Parent.Gender
	}

	if dcr.NewParentTelephone != nil && *dcr.NewParentTelephone != Parent.Telephone {
		comparedData.Telephone = *dcr.NewParentTelephone
	} else {
		comparedData.Telephone = Parent.Telephone
	}

	if dcr.NewParentEmail != nil && *dcr.NewParentEmail != *Parent.Email {
		comparedData.Email = dcr.NewParentEmail
	} else {
		comparedData.Email = Parent.Email
	}

	// Always update the timestamp
	comparedData.UpdatedAt = tNow

	// Check if parent is associated with any students
	err = tx.Where("parent_id = ?", Parent.ParentID).Find(&AssociatedStudent).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("parent doesn't associate with any student, error: %v", err)
	}

	var parentTelInStudent int64
	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ?", req["telephone"]).Count(&parentTelInStudent).Error
	if err != nil {
		return nil, fmt.Errorf("error checking parent telephone in student table: %v", err)
	}

	if parentTelInStudent > 0 {
		return nil, fmt.Errorf("parent with telephone %s already exist in student", comparedData.Telephone)
	}

	// Check for an existing parent record with the same details
	var ExistingParent domain.Parent
	err = tx.Where("(name = ? OR telephone = ? OR email = ?)", req["name"], req["telephone"], req["email"]).First(&ExistingParent).Error
	if err == gorm.ErrRecordNotFound {
		// If no existing parent found, update the current parent
		err = tx.Model(&domain.Parent{}).Where("parent_id = ?", Parent.ParentID).Updates(&comparedData).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update parent, error: %v", err)
		}
		err = spr.db.WithContext(ctx).Model(&domain.DataChangeRequest{}).Where("old_parent_telephone = ? AND is_reviewed IS FALSE", oldTelephone).Updates(&domain.DataChangeRequest{
			IsReviewed: true,
		}).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to review data change request, error: %v", err)
		}
		tx.Commit()
		return nil, nil
	}

	var msgs *string
	// Assign associated students to the existing parent
	for _, student := range AssociatedStudent {
		err = tx.Model(&domain.Student{}).Where("student_id = ? AND parent_id = ?", student.StudentID, Parent.ParentID).Updates(&domain.Student{
			ParentID: ExistingParent.ParentID,
		}).Error
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to assign student to existing parent, error: %v", err)
		}
		message := fmt.Sprintf("Parent data already exists, allocating %d students to the existing parent name: %s, telephone: %s, email: %s", len(AssociatedStudent), ExistingParent.Name, ExistingParent.Telephone, *ExistingParent.Email)
		msgs = &message
	}

	err = spr.db.WithContext(ctx).Model(&domain.DataChangeRequest{}).Where("old_parent_telephone = ? AND is_reviewed IS FALSE", oldTelephone).Updates(&domain.DataChangeRequest{
		IsReviewed: true,
	}).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to review data change request, error: %v", err)
	}

	if msgs != nil {
		tx.Commit()
		return msgs, nil
	}

	tx.Commit()
	return nil, nil
}

func (spr *studentParentRepository) CreateStudentAndParent(ctx context.Context, req *domain.StudentAndParent) (*string, *[]string) {
	var errList []string

	if req.Student.Telephone == req.Parent.Telephone {
		errList = append(errList, "Student and parent cant have the same telephone")
	}
	// ========================================STUDENT=======================================================
	// Validate NSN
	nsnLength := len(req.Student.NSN)
	if nsnLength > 10 {
		errList = append(errList, "NSN length exceeds maximum length of 10")
	}
	// Validate GradeLabel
	match, _ := regexp.MatchString("^[A-Za-z]+$", req.Student.GradeLabel)
	if !match {
		errList = append(errList, fmt.Sprintf("Invalid Grade Label: %s. Only letters (A-Z, a-z) are allowed.", req.Student.GradeLabel))
	}
	//  Validate telephone length bjir
	studTelLength := len(req.Student.Telephone)
	if studTelLength > 13 {
		errList = append(errList, "Student telephone should not be more than 13 number")
	}
	// Check for duplicate student telephone
	var studentCount int64
	err := spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ?", req.Student.Telephone).Count(&studentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student telephone: %v", err))
	} else if studentCount > 0 {
		errList = append(errList, fmt.Sprintf("Student with telephone %s already exists", req.Student.Telephone))
	}

	// Check for duplicate student telephone in parent table
	var studentCountInParent int64
	err = spr.db.WithContext(ctx).Model(&domain.Parent{}).Where("telephone = ?", req.Student.Telephone).Count(&studentCountInParent).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student telephone: %v", err))
	} else if studentCountInParent > 0 {
		errList = append(errList, fmt.Sprintf("Student with telephone %s already exists in parent", req.Student.Telephone))
	}

	var studentCountNSN int64
	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("nsn = ?", req.Student.NSN).Count(&studentCountNSN).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student nsn: %v", err))
	} else if studentCountNSN > 0 {
		errList = append(errList, fmt.Sprintf("Student with nsn %s already exists", req.Student.NSN))
	}

	// Check for duplicate student name
	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("name = ?", req.Student.Name).Count(&studentCount).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student name: %v", err))
	} else if studentCount > 0 {
		errList = append(errList, fmt.Sprintf("Student with name %s already exists", req.Student.Name))
	}
	// ========================================PARENT========================================================
	// Normalize Parent email if provided
	if req.Parent.Email != nil && *req.Parent.Email != "" {
		emailLowered := strings.ToLower(strings.TrimSpace(*req.Parent.Email))
		req.Parent.Email = &emailLowered

		// Validate email format
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		match, _ := regexp.MatchString(emailRegex, *req.Parent.Email)
		if !match {
			errList = append(errList, fmt.Sprintf("Invalid email format for parent: %s", *req.Parent.Email))
		}
	}
	// Validate parent telephone length
	parTelLength := len(req.Parent.Telephone)
	if parTelLength > 13 {
		errList = append(errList, "Parent telephone should not be more than 13 number")
	}

	var parentTelInStudent int64
	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ?", req.Parent.Telephone).Count(&parentTelInStudent).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking parent telephone in student table: %v", err))
	}

	if parentTelInStudent > 0 {
		errList = append(errList, fmt.Sprintf("Parent with telephone %s already exist in student", req.Parent.Telephone))
	}

	// If errors exist, return immediately
	if len(errList) > 0 {
		return nil, &errList
	}

	// Find existing parent with matching attributes
	var existingParent domain.Parent
	err = spr.db.WithContext(ctx).Where(
		"(name = ? OR telephone = ? OR email = ?)",
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
		return nil, &[]string{fmt.Sprintf("Could not begin transaction: %v", tx.Error)}
	}
	var msgs *string
	if oneToManyCondition {
		// Parent exists, create student with reference
		req.Student.ParentID = existingParent.ParentID
		req.Student.CreatedAt = time.Now()
		req.Student.UpdatedAt = req.Student.CreatedAt
		req.Student.GradeLabel = strings.ToUpper(req.Student.GradeLabel)

		if err := tx.WithContext(ctx).Create(&req.Student).Error; err != nil {
			tx.Rollback()
			return nil, &[]string{fmt.Sprintf("Could not insert student: %v", err)}
		}

		message := fmt.Sprintf(
			"Parent data already exists, allocating the student to the existing parent name: %s, telephone: %s, email: %s",
			existingParent.Name,
			existingParent.Telephone,
			func() string {
				if existingParent.Email == nil {
					return "N/A"
				}
				return *existingParent.Email
			}(),
		)
		msgs = &message
	} else {
		// Create new parent
		req.Parent.CreatedAt = time.Now()
		req.Parent.UpdatedAt = req.Parent.CreatedAt

		if err := tx.WithContext(ctx).Create(&req.Parent).Error; err != nil {
			tx.Rollback()
			return nil, &[]string{fmt.Sprintf("Could not insert parent: %v", err)}
		}

		// Create new student
		req.Student.ParentID = req.Parent.ParentID
		req.Student.CreatedAt = time.Now()
		req.Student.UpdatedAt = req.Student.CreatedAt
		req.Student.GradeLabel = strings.ToUpper(req.Student.GradeLabel)

		if err := tx.WithContext(ctx).Create(&req.Student).Error; err != nil {
			tx.Rollback()
			return nil, &[]string{fmt.Sprintf("Could not insert student: %v", err)}
		}
	}

	if msgs != nil {
		tx.Commit()
		return msgs, nil
	}

	if err := tx.Commit().Error; err != nil {
		return nil, &[]string{fmt.Sprintf("Could not commit transaction: %v", err)}
	}

	return nil, nil
}

func (spr *studentParentRepository) ImportCSV(ctx context.Context, payload *[]domain.StudentAndParent) (*[]string, error) {
	var duplicateMessages []string
	var validRecords []domain.StudentAndParent
	now := time.Now()

	// Validate and filter the records
	for index, record := range *payload {
		isDuplicate := false

		// Validate Parent Telephone
		if len(record.Parent.Telephone) > 13 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: parent telephone %s exceeds max length (13)", index+2, record.Parent.Telephone))
			isDuplicate = true
		}

		// Validate Student Telephone
		var studentExists domain.Student
		if len(record.Student.Telephone) > 13 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student telephone %s exceeds max length (13)", index+2, record.Student.Telephone))
			isDuplicate = true
		} else if err := spr.db.WithContext(ctx).Where("telephone = ?", record.Student.Telephone).First(&studentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student telephone %s already exists", index+2, record.Student.Telephone))
			isDuplicate = true
		}

		// Validate parent telephone (checking availablity parent telephone in student)
		var parentTelInStudent int64
		err := spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ?", record.Parent.Telephone).Count(&parentTelInStudent).Error
		if err != nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("Error checking parent telephone in student table: %v", err))
			isDuplicate = true
		}

		if parentTelInStudent > 0 {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("Parent with telephone %s already exist in student", record.Parent.Telephone))
			isDuplicate = true
		}

		// Validate Student Name
		if err := spr.db.WithContext(ctx).Where("name = ?", record.Student.Name).First(&studentExists).Error; err == nil {
			duplicateMessages = append(duplicateMessages, fmt.Sprintf("row %d: student name %s already exists", index+2, record.Student.Name))
			isDuplicate = true
		}

		// Skip records with validation errors
		if isDuplicate {
			continue
		}

		// Assign timestamps
		record.Parent.CreatedAt = now
		record.Parent.UpdatedAt = now
		record.Student.CreatedAt = now
		record.Student.UpdatedAt = now

		// Add valid records
		validRecords = append(validRecords, record)
	}

	// Return duplicate messages if any
	if len(duplicateMessages) > 0 {
		return &duplicateMessages, nil
	}

	// Insert valid records into the database
	err := spr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, record := range validRecords {
			var parentExist domain.Parent

			// Check if parent already exists
			err := tx.Where("(name = ? OR telephone = ? OR email = ?)",
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
				// Parent does not exist, create a new one
				if err := tx.Create(&record.Parent).Error; err != nil {
					return fmt.Errorf("failed to insert parent: %w", err)
				}
				record.Student.ParentID = record.Parent.ParentID
			} else {
				// Use the existing parent's ID
				record.Student.ParentID = parentExist.ParentID
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

func (spr *studentParentRepository) UpdateStudentAndParent(ctx context.Context, id int, req *domain.StudentAndParent) (*string, *[]string) {
	var student domain.Student
	var errList []string
	// ========================================STUDENT=======================================================
	// Validate GradeLabel to only contain letters
	match, _ := regexp.MatchString("^[A-Za-z]+$", req.Student.GradeLabel)
	if !match {
		errList = append(errList, fmt.Sprintf("Invalid Grade Label: %s. Only letters (A-Z, a-z) are allowed.", req.Student.GradeLabel))
	}
	req.Student.GradeLabel = strings.ToUpper(req.Student.GradeLabel)

	studTelLength := len(req.Student.Telephone)
	if studTelLength > 13 {
		errList = append(errList, "Student telephone should not be more than 13 number")
	}

	// ========================================PARENT=======================================================
	// Normalize email if provided
	if req.Parent.Email != nil && *req.Parent.Email != "" {
		emailLowered := strings.ToLower(strings.TrimSpace(*req.Parent.Email))
		req.Parent.Email = &emailLowered

		// Validate email format
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		match, _ := regexp.MatchString(emailRegex, *req.Parent.Email)
		if !match {
			errList = append(errList, fmt.Sprintf("Invalid email format for parent: %s", *req.Parent.Email))
		}
	}

	parTelLength := len(req.Parent.Telephone)
	if parTelLength > 13 {
		errList = append(errList, "Parent telephone should not be more than 13 number")
	}

	// Start a transaction
	tx := spr.db.Begin()
	if err := tx.Error; err != nil {
		errList = append(errList, fmt.Sprintf("could not begin transaction: %v", err))
		return nil, &errList
	}

	now := time.Now()
	req.Student.UpdatedAt = now
	req.Parent.UpdatedAt = now

	// =======================================STUDENT=======================================================
	nsnLength := len(req.Student.NSN)
	if nsnLength > 10 {
		errList = append(errList, "NSN length exceeds maximum length of 10")
	}

	var studentCountNSN int64
	err := spr.db.WithContext(ctx).Model(&domain.Student{}).Where("nsn = ? AND student_id != ?", req.Student.NSN, id).Count(&studentCountNSN).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student nsn: %v", err))
	} else if studentCountNSN > 0 {
		errList = append(errList, fmt.Sprintf("Student with nsn %s already exists", req.Student.NSN))
	}

	if req.Student.Name != "" {
		err := spr.db.WithContext(ctx).
			Where("name = ? AND student_id != ?", req.Student.Name, id).
			First(&domain.Student{}).Error
		if err == nil {
			errList = append(errList, fmt.Sprintf("Student with name %s already exists", req.Student.Name))
		}
	}
	if req.Student.Telephone != "" {
		err := spr.db.WithContext(ctx).
			Where("telephone = ? AND student_id != ?", req.Student.Telephone, id).
			First(&domain.Student{}).Error
		if err == nil {
			errList = append(errList, fmt.Sprintf("Student with telephone %s already exists", req.Student.Telephone))
		}
	}

	var studTelInParent int64
	err = spr.db.WithContext(ctx).Model(&domain.Parent{}).Where("telephone = ?", req.Student.Telephone).Count(&studTelInParent).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking for student telephone: %v", err))
	} else if studTelInParent > 0 {
		errList = append(errList, fmt.Sprintf("Student with telephone %s already exists in parent", req.Student.Telephone))
	}

	// ========================================PARENT=======================================================
	var parentTelInStudent int64
	err = spr.db.WithContext(ctx).Model(&domain.Student{}).Where("telephone = ?", req.Parent.Telephone).Count(&parentTelInStudent).Error
	if err != nil {
		errList = append(errList, fmt.Sprintf("Error checking parent telephone in student table: %v", err))
	}

	if parentTelInStudent > 0 {
		errList = append(errList, fmt.Sprintf("Parent with telephone %s already exist in student", req.Parent.Telephone))
	}

	if len(errList) > 0 {
		tx.Rollback()
		return nil, &errList
	}

	// Fetch existing student with parent
	err = spr.db.WithContext(ctx).Preload("Parent").Where("student_id = ?", id).First(&student).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errList = append(errList, fmt.Sprintf("can't find student with id %d", id))
			return nil, &errList
		}
		errList = append(errList, fmt.Sprintf("database error: %v", err))
		return nil, &errList
	}

	// Build map of updated fields for student
	updatedStudentFields := make(map[string]interface{})
	if req.Student.Name != "" && req.Student.Name != student.Name {
		updatedStudentFields["name"] = req.Student.Name
	}
	if req.Student.NSN != "" && req.Student.NSN != student.NSN {
		updatedStudentFields["nsn"] = req.Student.NSN
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

	var msgs *string
	if len(updatedParentFields) > 0 {
		var existingParent domain.Parent
		err = tx.WithContext(ctx).Where(
			"(name = ? OR telephone = ? OR email = ?)",
			updatedParentFields["name"], updatedParentFields["telephone"], updatedParentFields["email"],
		).First(&existingParent).Error

		if err == nil {
			// Update parent_id to the existing parent's ID
			updatedStudentFields["ParentID"] = existingParent.ParentID
			message := fmt.Sprintf("Parent data already exists, allocating the student to the existing parent named: %s", existingParent.Name)
			msgs = &message

		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// Update the parent and keep the current parent_id
			if err := tx.WithContext(ctx).
				Model(domain.Parent{}).
				Where("parent_id = ?", student.ParentID).
				Updates(updatedParentFields).Error; err != nil {
				tx.Rollback()
				errList = append(errList, fmt.Sprintf("failed to update parent: %v", err))
				return nil, &errList
			}
		} else {
			tx.Rollback()
			errList = append(errList, fmt.Sprintf("database error while checking parent: %v", err))
			return nil, &errList
		}
	}

	err = tx.WithContext(ctx).
		Model(domain.Student{}).
		Where("student_id = ?", student.StudentID).
		Updates(updatedStudentFields).Error
	if err != nil {
		tx.Rollback()
		errList = append(errList, fmt.Sprintf("failed to update student: %v", err))
		return nil, &errList
	}

	if msgs != nil {
		tx.Commit()
		return msgs, nil
	}

	if err := tx.Commit().Error; err != nil {
		errList = append(errList, fmt.Sprintf("could not commit transaction: %v", err))
		return nil, &errList
	}

	return nil, nil
}

func (spr *studentParentRepository) SPMassDelete(ctx context.Context, studentIDs *[]int) error {
	// Start a transaction
	// tx := spr.db.Begin()
	// if err := tx.Error; err != nil {
	// 	return fmt.Errorf("could not begin transaction: %v", err)
	// }

	// currentTime := time.Now()

	// // Iterate over the student IDs to process each student
	// for _, studentID := range *studentIDs {
	// 	// Retrieve the parent_id from the student record
	// 	var student domain.Student
	// 	err := tx.WithContext(ctx).
	// 		Where("student_id = ?", studentID).
	// 		First(&student).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		if err == gorm.ErrRecordNotFound {
	// 			return fmt.Errorf("student with ID %d not found", studentID)
	// 		}
	// 		return fmt.Errorf("error retrieving student: %v", err)
	// 	}

	// 	// Count remaining active students associated with the same parent
	// 	var remainingStudentCount int64
	// 	err = tx.WithContext(ctx).
	// 		Model(&domain.Student{}).
	// 		Where("parent_id = ?", student.ParentID).
	// 		Count(&remainingStudentCount).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return fmt.Errorf("error counting remaining students: %v", err)
	// 	}

	// 	if remainingStudentCount > 1 {
	// 		// Soft delete only the student by setting DeletedAt
	// 		err = tx.WithContext(ctx).
	// 			Model(&domain.Student{}).
	// 			Where("student_id = ?", studentID).
	// 			Update("deleted_at", currentTime).Error
	// 		if err != nil {
	// 			tx.Rollback()
	// 			return fmt.Errorf("error soft deleting student: %v", err)
	// 		}
	// 	} else {
	// 		// Soft delete both the student and the parent by setting DeletedAt
	// 		err = tx.WithContext(ctx).
	// 			Model(&domain.Student{}).
	// 			Where("student_id = ?", studentID).
	// 			Update("deleted_at", currentTime).Error
	// 		if err != nil {
	// 			tx.Rollback()
	// 			return fmt.Errorf("error soft deleting student: %v", err)
	// 		}

	// 		err = tx.WithContext(ctx).
	// 			Model(&domain.Parent{}).
	// 			Where("parent_id = ?", student.ParentID).
	// 			Update("deleted_at", currentTime).Error
	// 		if err != nil {
	// 			tx.Rollback()
	// 			return fmt.Errorf("error soft deleting parent: %v", err)
	// 		}
	// 	}
	// }

	// // Commit the transaction
	// err := tx.Commit().Error
	// if err != nil {
	// 	return fmt.Errorf("could not commit transaction: %v", err)
	// }

	return nil
}

func (spr *studentParentRepository) DeleteStudentAndParent(ctx context.Context, studentID int) error {
	// // Start a transaction
	// tx := spr.db.Begin()
	// if err := tx.Error; err != nil {
	// 	return fmt.Errorf("could not begin transaction: %v", err)
	// }

	// // Retrieve the parent_id from the student record
	// var student domain.Student
	// err := tx.WithContext(ctx).
	// 	Select("parent_id").
	// 	Where("student_id = ?", studentID).
	// 	First(&student).Error
	// if err != nil {
	// 	tx.Rollback()
	// 	if err == gorm.ErrRecordNotFound {
	// 		return fmt.Errorf("student with ID %d not found", studentID)
	// 	}
	// 	return fmt.Errorf("error retrieving student: %v", err)
	// }

	// // Count students associated with the same parent
	// var studentCount int64
	// err = tx.WithContext(ctx).
	// 	Model(&domain.Student{}).
	// 	Where("parent_id = ?", student.ParentID).
	// 	Count(&studentCount).Error
	// if err != nil {
	// 	tx.Rollback()
	// 	return fmt.Errorf("error counting students: %v", err)
	// }

	// currentTime := time.Now()

	// if studentCount > 1 {
	// 	// Soft delete only the student by setting DeletedAt
	// 	err = tx.WithContext(ctx).
	// 		Model(&domain.Student{}).
	// 		Where("student_id = ?", studentID).
	// 		Update("deleted_at", currentTime).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return fmt.Errorf("error soft deleting student: %v", err)
	// 	}
	// } else {
	// 	// Soft delete both the student and the parent by setting DeletedAt
	// 	err = tx.WithContext(ctx).
	// 		Model(&domain.Student{}).
	// 		Where("student_id = ?", studentID).
	// 		Update("deleted_at", currentTime).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return fmt.Errorf("error soft deleting student: %v", err)
	// 	}

	// 	err = tx.WithContext(ctx).
	// 		Model(&domain.Parent{}).
	// 		Where("parent_id = ?", student.ParentID).
	// 		Update("deleted_at", currentTime).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return fmt.Errorf("error soft deleting parent: %v", err)
	// 	}
	// }

	// // Commit the transaction
	// err = tx.Commit().Error
	// if err != nil {
	// 	return fmt.Errorf("could not commit transaction: %v", err)
	// }

	return nil
}

func (spr *studentParentRepository) DeleteStudentAndParentMass(ctx context.Context, studentIDs *[]int) error {
	// // Start a transaction
	// tx := spr.db.Begin()
	// if err := tx.Error; err != nil {
	// 	return fmt.Errorf("could not begin transaction: %v", err)
	// }

	// currentTime := time.Now()

	// // Iterate over the student IDs to process each student
	// for _, studentID := range *studentIDs {
	// 	// Retrieve the parent_id from the student record
	// 	var student domain.Student
	// 	err := tx.WithContext(ctx).
	// 		Where("student_id = ?", studentID).
	// 		First(&student).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		if err == gorm.ErrRecordNotFound {
	// 			return fmt.Errorf("student with ID %d not found", studentID)
	// 		}
	// 		return fmt.Errorf("error retrieving student: %v", err)
	// 	}

	// 	// Count remaining active students associated with the same parent
	// 	var remainingStudentCount int64
	// 	err = tx.WithContext(ctx).
	// 		Model(&domain.Student{}).
	// 		Where("parent_id = ?", student.ParentID).
	// 		Count(&remainingStudentCount).Error
	// 	if err != nil {
	// 		tx.Rollback()
	// 		return fmt.Errorf("error counting remaining students: %v", err)
	// 	}

	// 	if remainingStudentCount > 1 {
	// 		// Soft delete only the student by setting DeletedAt
	// 		err = tx.WithContext(ctx).
	// 			Model(&domain.Student{}).
	// 			Where("student_id = ?", studentID).
	// 			Update("deleted_at", currentTime).Error
	// 		if err != nil {
	// 			tx.Rollback()
	// 			return fmt.Errorf("error soft deleting student: %v", err)
	// 		}
	// 	} else {
	// 		// Soft delete both the student and the parent by setting DeletedAt
	// 		err = tx.WithContext(ctx).
	// 			Model(&domain.Student{}).
	// 			Where("student_id = ?", studentID).
	// 			Update("deleted_at", currentTime).Error
	// 		if err != nil {
	// 			tx.Rollback()
	// 			return fmt.Errorf("error soft deleting student: %v", err)
	// 		}

	// 		err = tx.WithContext(ctx).
	// 			Model(&domain.Parent{}).
	// 			Where("parent_id = ?", student.ParentID).
	// 			Update("deleted_at", currentTime).Error
	// 		if err != nil {
	// 			tx.Rollback()
	// 			return fmt.Errorf("error soft deleting parent: %v", err)
	// 		}
	// 	}
	// }

	// // Commit the transaction
	// err := tx.Commit().Error
	// if err != nil {
	// 	return fmt.Errorf("could not commit transaction: %v", err)
	// }

	return nil
}

func (spr *studentParentRepository) GetStudentDetailsByID(ctx context.Context, studentID int) (*domain.StudentAndParent, error) {
	var result domain.StudentAndParent
	err := spr.db.WithContext(ctx).
		Preload("Parent").
		Where("students.student_id = ?", studentID).
		First(&result.Student).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("student with ID %d not found", studentID)
		}
		return nil, fmt.Errorf("could not fetch student details: %v", err)
	}

	// If ParentID is 0, it means the student has no associated parent
	if result.Student.ParentID == 0 {
		result.Student.Parent = domain.Parent{} // Explicitly set to empty struct
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
	var countVariable int64
	var parentCount int64
	err := spr.db.WithContext(ctx).Model(&domain.Parent{}).Where("telephone = ?", datas.OldParentTelephone).Count(&parentCount).Error
	if err != nil {
		return err
	}
	if parentCount == 0 {
		return fmt.Errorf("parent with telephone %s does not exist or registered", datas.OldParentTelephone)
	}

	err = spr.db.WithContext(ctx).Model(&domain.DataChangeRequest{}).Where("old_parent_telephone = ? AND is_reviewed IS FALSE", datas.OldParentTelephone).Count(&countVariable).Error
	if err != nil {
		return err
	}

	if countVariable > 0 {
		return fmt.Errorf("data change request with old telephone parent: %s already exists and has not been reviewed yet. If this is urgent, please contact the school directly", datas.OldParentTelephone)
	}

	err = spr.db.WithContext(ctx).Create(&datas).Error
	if err != nil {
		return err
	}

	return nil
}

func (spr *studentParentRepository) DeleteDCR(ctx context.Context, dcrID int) error {
	result := spr.db.WithContext(ctx).
		Model(&domain.DataChangeRequest{}).
		Where("request_id = ?", dcrID).
		Update("deleted_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("failed to delete for request_id %d: %w", dcrID, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no data change request found with request id %d", dcrID)
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
