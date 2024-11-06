package delivery

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"notification/config"
	"notification/domain"
	"notification/middleware"
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

	route := app.Group("/student-and-parent")
	route.Post("/insert", handler.CreateStudentAndParent)
	route.Post("/import", handler.UploadAndImport)
	route.Put("/modify/:id", handler.UpdateStudentAndParent)
	route.Delete("/rm/:id", handler.DeleteStudentAndParent)
	route.Get("/student/:id", handler.GetStudentDetailsByID)
}

func NewStudentParentHandlerDeploy(app *fiber.App, useCase domain.StudentParentUseCase) {
	handler := &studentParentHandler{
		uc: useCase,
	}

	route := app.Group("/student-and-parent")
	route.Post("/insert", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateStudentAndParent)
	route.Post("/import", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.UploadAndImport)
	route.Put("/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.UpdateStudentAndParent)
	route.Delete("/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStudentAndParent)
	route.Get("/student/:id", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.GetStudentDetailsByID)
}

func (sph *studentParentHandler) CreateStudentAndParent(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "CreateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid request body",
		})
	}

	_, err := govalidator.ValidateStruct(req.Student)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "CreateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid Student request body",
		})
	}

	_, err = govalidator.ValidateStruct(req.Parent)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "CreateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid Parent request body",
		})
	}

	if req.Parent.Email != nil && *req.Parent.Email == "" {
		req.Parent.Email = nil
	}

	if err := sph.uc.CreateStudentAndParentUC(c.Context(), &req); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "CreateStudentAndParent")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err,
			"message": "Failed to Create Student and Parent",
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusCreated, "CreateStudentAndParent")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent created successfully",
	})
}

func (sph *studentParentHandler) UploadAndImport(c *fiber.Ctx) error {
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
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
				"message": "Failed to create upload directory",
			})
		}
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

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":            false,
			"error":              err.Error(),
			"message":            "Failed to process CSV file",
			"invalid_telephones": invalidTelephones,
		})
	}

	if resDupe != nil && len(*resDupe) > 0 {
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
	if sph.uc == nil {
		return nil, nil, fmt.Errorf("use case service is not initialized")
	}

	var listStudentAndParent []domain.StudentAndParent
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

	// Start from row 2 because row 1 is the header
	for i, row := range records[1:] {
		if len(row) < 8 {
			log.Printf("Skipping row %d due to insufficient columns", i+2)
			continue
		}

		studentDataHolder = domain.Student{
			Name:      row[0],
			Class:     row[1],
			Gender:    row[2],
			Telephone: row[3],
			ParentID:  0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = govalidator.ValidateStruct(studentDataHolder)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d: error validating student: %v", i+2, err)
		}

		parentDataHolder = domain.Parent{
			Name:      row[4],
			Gender:    row[5],
			Telephone: row[6],
			Email:     getStringPointer(row[7]),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = govalidator.ValidateStruct(parentDataHolder)
		if err != nil {
			log.Printf("Parent validation failed on row %d: %v", i+2, err)
			return nil, nil, fmt.Errorf("row %d: error validating parent: %v", i+2, err)
		}

		studNParent := domain.StudentAndParent{
			Student: studentDataHolder,
			Parent:  parentDataHolder,
		}
		// Append to the list
		listStudentAndParent = append(listStudentAndParent, studNParent)
	}

	duplicates, err := sph.uc.ImportCSV(c, &listStudentAndParent)
	if err != nil {
		return nil, nil, fmt.Errorf("error importing CSV data: %v", err)
	}

	if duplicates != nil && len(*duplicates) > 0 {
		return duplicates, nil, nil
	}

	return nil, nil, nil
}

// Helper function to get a pointer to a string
func getStringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (sph *studentParentHandler) UpdateStudentAndParent(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID Required",
		})
	}

	convertetID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid ID",
			"error":   err.Error(),
		})
	}

	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid request body",
		})
	}

	_, err = govalidator.ValidateStruct(&req)
	if err != nil {
		validationErrors := govalidator.ErrorsByField(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"errors":  validationErrors,
			"message": "Invalid request body",
		})
	}

	errList := sph.uc.UpdateStudentAndParent(c.Context(), convertetID, &req)
	if errList != nil && len(*errList) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   *errList,
			"message": "Failed to update student and parent",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent updated successfully",
	})
}

func (sph *studentParentHandler) DeleteStudentAndParent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	if err := sph.uc.DeleteStudentAndParent(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete student",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student deleted successfully",
	})
}

func (sph *studentParentHandler) GetStudentDetailsByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	student, err := sph.uc.GetStudentDetailsByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get student",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent retrieved successfully",
		"data":    student.Student,
	})
}

func (sph *studentParentHandler) DataChangeRequest(c *fiber.Ctx) error {
	var datas domain.DataChangeRequest

	err := c.BodyParser(&datas)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Request",
			"error":   err,
		})
	}

	err = sph.uc.DataChangeRequest(c.Context(), datas)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to send data change request",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully sent data changes",
	})
}
