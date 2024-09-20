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
	route.Put("/modify/:id", handler.UpdateStudentAndParent)
}

func (sph *studentParentHandler) CreateStudentAndParent(c *fiber.Ctx) error {
	wg.Add(1)
	defer wg.Done()

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

	if err := sph.uc.CreateStudentAndParentUC(c.Context(), &req); err != nil {
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
	wg.Add(1)
	defer wg.Done()
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
	uploadDir := "./uploads"
	// Ensure upload directory exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	// Save the file
	filePath := filepath.Join(uploadDir, file.Filename)
	err = c.SaveFile(file, filePath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to save file",
		})
	}

	// Process the CSV file and get duplicate records
	resDupe, invalidTelephones, err := sph.processCSVFile(c.Context(), filePath)

	if invalidTelephones != nil && err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":            false,
			"error":              err.Error(),
			"message":            "Failed to process CSV file",
			"invalid_telephones": invalidTelephones,
		})
	}

	if resDupe != nil && err == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success":    false,
			"message":    "File processed successfully, but some duplicates were found.",
			"duplicates": resDupe,
		})
	}

	// If no errors and no duplicates, return success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "File processed successfully",
	})

}

func (sph *studentParentHandler) processCSVFile(c context.Context, filePath string) (*[]string, *[]string, error) {
	var listStudentAndParent []domain.StudentAndParent
	var invalidTelephoneType []string
	var parentDataHolder domain.Parent
	var studentDataHolder domain.Student

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()
	defer func() {
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete file: %v", err)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	// Assume the first row contains headers, so start processing from row 2
	for i, row := range records[1:] {
		if len(row) < 8 { // Ensure enough columns
			log.Printf("Skipping row %d due to insufficient columns", i+2)
			continue
		}

		convertStudTelephone, errStudentConvert := strconv.Atoi(row[3])
		if errStudentConvert != nil {
			txt := fmt.Sprintf("Student telephone should be a number, Found : %s", row[3])
			invalidTelephoneType = append(invalidTelephoneType, txt)
		}

		if errStudentConvert == nil {
			// Process student data
			studentDataHolder = domain.Student{
				Name:      row[0],
				Class:     row[1],
				Gender:    row[2],
				Telephone: convertStudTelephone,
				ParentID:  0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err = govalidator.ValidateStruct(studentDataHolder)
			if err != nil {
				return nil, nil, fmt.Errorf("row %d: error validating student: %v", i+2, err)
			}

		}

		convertParentTelephone, errParentConvert := strconv.Atoi(row[6])
		if errParentConvert != nil {
			txt := fmt.Sprintf("Parent telephone should be a number, Found : %s", row[6])
			invalidTelephoneType = append(invalidTelephoneType, txt)
		}

		if errParentConvert == nil {
			// Process parent data
			parentDataHolder = domain.Parent{
				Name:      row[4],
				Gender:    row[5],
				Telephone: convertParentTelephone,
				Email:     &row[7],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err = govalidator.ValidateStruct(parentDataHolder)
			if err != nil {
				return nil, nil, fmt.Errorf("row %d: error validating parent: %v", i+2, err)
			}
		}

		if errStudentConvert == nil && errParentConvert == nil {
			// Combine student and parent into a single struct
			studNParent := domain.StudentAndParent{
				Student: studentDataHolder,
				Parent:  parentDataHolder,
			}
			// Append to the list
			listStudentAndParent = append(listStudentAndParent, studNParent)
		}
	}

	if len(invalidTelephoneType) > 0 {
		return nil, &invalidTelephoneType, fmt.Errorf("Invalid type of telephones found")
	}

	// Use case logic for importing students and parents in bulk
	duplicates, err := sph.uc.ImportCSV(c, &listStudentAndParent)
	if err != nil {
		return nil, nil, fmt.Errorf("error importing CSV data: %v", err)
	}

	if len(*duplicates) > 0 {
		return duplicates, nil, nil
	}

	return nil, nil, nil

}

func (sph *studentParentHandler) UpdateStudentAndParent(c *fiber.Ctx) error {
	// Get the ID from the URL parameters
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID is required",
		})
	}

	convertetID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err,
		})
	}

	// Parse the request body to get the student and parent data
	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid request body",
		})
	}

	// Validate Student
	if req.Student.Name != "" || req.Student.Telephone != 0 || req.Student.Class != "" || req.Student.Gender != "" {
		if _, err := govalidator.ValidateStruct(req.Student); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
				"message": "Invalid student data",
			})
		}
	}

	// Validate Parent
	if req.Parent.Name != "" || req.Parent.Telephone != 0 || req.Parent.Gender != "" || req.Parent.Email != nil {
		if _, err := govalidator.ValidateStruct(req.Parent); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
				"message": "Invalid parent data",
			})
		}
	}

	if err := sph.uc.UpdateStudentAndParent(c.Context(), convertetID, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to update student and parent",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent updated successfully",
	})
}
