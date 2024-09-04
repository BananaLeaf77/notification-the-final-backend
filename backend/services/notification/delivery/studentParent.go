package delivery

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"notification/domain"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
)

type studentParentHandler struct {
	uc domain.StudentParentUseCase
}

func NewStudentParentHandler(app *fiber.App, useCase domain.StudentParentUseCase) {
	handler := &studentParentHandler{
		uc: useCase,
	}

	route := app.Group("/student_and_parent")
	route.Post("/insert", handler.CreateStudentAndParent)
	route.Post("/import", handler.UploadAndImport)
}

func (sph *studentParentHandler) CreateStudentAndParent(c *fiber.Ctx) error {
	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid request body",
		})
	}

	// Validate Student
	if _, err := govalidator.ValidateStruct(req.Student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid student data",
		})
	}

	// Validate Parent
	if _, err := govalidator.ValidateStruct(req.Parent); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid parent data",
		})
	}

	ctx := context.Background()
	if err := sph.uc.CreateStudentAndParentUC(ctx, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Internal Server Error",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent created successfully",
	})
}

func (sph *studentParentHandler) UploadAndImport(c *fiber.Ctx) error {
	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to parse file",
		})
	}

	// Define upload directory
	uploadDir := "../uploads"
	// Ensure upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	// Save the file
	filePath := filepath.Join(uploadDir, file.Filename)
	err = c.SaveFile(file, filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to save file",
		})
	}

	// Process the CSV file and get duplicate records
	duplicates, err := sph.processCSVFile(filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to process CSV file",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":    true,
		"message":    "File processed successfully",
		"duplicates": duplicates, // Include duplicate information in the response
	})
}

func (sph *studentParentHandler) processCSVFile(filePath string) ([]string, error) {
	var listStudentAndParent []domain.StudentAndParent
	var duplicateMessages []string // Store messages about duplicates
	var duplicateParentTelephones []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	// Assume the first row contains headers, so start processing from row 2
	for i, row := range records[1:] {
		if len(row) < 8 { // Ensure enough columns
			log.Printf("Skipping row %d due to insufficient columns", i+2)
			continue
		}

		convertStudTelephone, err := strconv.Atoi(row[3])
		if err != nil {
			txt := fmt.Sprintf("Telephone should be number, Found : %s", row[3])
			duplicateParentTelephones = append(duplicateParentTelephones, txt)
		}

		// Process student data
		student := domain.Student{
			Name:      row[0],
			Class:     row[1],
			Gender:    row[2],
			Telephone: convertStudTelephone,
			ParentID:  0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = govalidator.ValidateStruct(student)
		if err != nil {
			return nil, fmt.Errorf("row %d: error validating student: %v", i+2, err)
		}

		convertParentTelephone, err := strconv.Atoi(row[6])
		if err != nil {
			return nil, fmt.Errorf("Telephone should be number, Found : %s", row[6])
		}
		// Process parent data
		parent := domain.Parent{
			Name:      row[4],
			Gender:    row[5],
			Telephone: convertParentTelephone,
			Email:     row[7],
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = govalidator.ValidateStruct(parent)
		if err != nil {
			return nil, fmt.Errorf("row %d: error validating parent: %v", i+2, err)
		}

		// Combine student and parent into a single struct
		studNParent := domain.StudentAndParent{
			Student: student,
			Parent:  parent,
		}

		// Append to the list
		listStudentAndParent = append(listStudentAndParent, studNParent)
		
	}
	
	if len(duplicateParentTelephones) > 0 {
		return duplicateParentTelephones, fmt.Errorf("Found parent duplicate telephone numbers")
	}

	// Use case logic for importing students and parents in bulk
	ctx := context.Background()
	duplicates, err := sph.uc.ImportCSV(ctx, &listStudentAndParent)
	if err != nil {
		return nil, fmt.Errorf("error importing CSV data: %v", err)
	}

	// Collect duplicates messages
	if len(*duplicates) > 0 {
		duplicateMessages = append(duplicateMessages, *duplicates...)
	}

	return duplicateMessages, err
}
