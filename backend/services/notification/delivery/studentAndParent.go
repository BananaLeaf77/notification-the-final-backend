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
	"regexp"
	"strconv"
	"strings"
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
	route.Post("/req/data-change-request", handler.DataChangeRequest)
	route.Get("/get-all-data-change-request", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllDataChangeRequest)
	route.Post("/rms", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.SPMassDelete)
	route.Get("/download-template", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DownloadTemplate)
}

func (sph *studentParentHandler) DownloadTemplate(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	filePath := "./template/up2.csv"

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="up2.csv"`)
	c.Set(fiber.HeaderContentType, "text/csv")

	err := c.SendFile(filePath, true)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DownloadTemplate")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to download template: " + err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DownloadTemplate")
	return nil
}

func (sph *studentParentHandler) GetAllDataChangeRequest(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	v, err := sph.uc.GetAllDataChangeRequest(c.Context())
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetAllDataChangeRequest")
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to get all data change request",
			"data":    nil,
		}))
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetAllDataChangeRequest")
	return c.Status(fiber.StatusOK).JSON((fiber.Map{
		"success": true,
		"message": "Data Change Request Retrieved Successfully",
		"data":    v,
	}))

}

func (sph *studentParentHandler) SPMassDelete(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	var payload struct {
		IDS []int `json:"student_ids"`
	}

	err := c.BodyParser(&payload)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "SPMassDelete")
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to delete users",
		}))
	}

	err = sph.uc.SPMassDelete(c.Context(), &payload.IDS)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "SPMassDelete")
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to delete users",
		}))
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "SPMassDelete")
	return c.Status(fiber.StatusOK).JSON((fiber.Map{
		"success": true,
		"message": "Users deleted successfully",
	}))
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

	if req.Student.GradeLabel == "" {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "CreateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid Grade Label",
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

	if req.Parent.Email != nil {
		trimmedEmail := strings.TrimSpace(*req.Parent.Email)
		if trimmedEmail == "" {
			req.Parent.Email = nil
		} else {
			req.Parent.Email = &trimmedEmail
		}
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
	userToken, _ := c.Locals("user").(*domain.Claims)

	file, err := c.FormFile("file")
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UploadAndImport")
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
			config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "UploadAndImport")
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
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UploadAndImport")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to save file",
		})
	}

	// Process the CSV file and get duplicate records
	badRequests, internalServerResponse, _ := sph.processCSVFile(c.Context(), filePath)

	if badRequests != nil && len(*badRequests) > 0 {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UploadAndImport")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": false,
			"message": "Import Failure, bad input found.",
			"error":   badRequests,
		})
	}

	if internalServerResponse != nil && len(*internalServerResponse) > 0 {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "UploadAndImport")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": false,
			"message": "Import Failure, duplicates found.",
			"error":   internalServerResponse,
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "UploadAndImport")
	// If no errors and no duplicates, return success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "File processed successfully",
	})
}

func (sph *studentParentHandler) processCSVFile(c context.Context, filePath string) (*[]string, *[]string, error) {
	var errList []string
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

	// Precompile the regular expression outside the loop
	gradeLabelRegex, err := regexp.Compile("^[A-Za-z]+$")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile regular expression: %v", err)
	}

	// Start from row 2 because row 1 is the header
	for i, row := range records[1:] {
		if len(row) < 8 {
			log.Printf("Skipping row %d due to insufficient columns", i+2)
			continue
		}
		// conv row grade
		convertedGrade, err := strconv.Atoi(row[1])
		if err != nil {
			errList = append(errList, fmt.Sprintf("row %d: grade should be number: %s", i+2, row[1]))
		}

		match := gradeLabelRegex.MatchString(row[2])
		if !match {
			errList = append(errList, fmt.Sprintf("row %d: Invalid Grade Label: %s. Only letters (A-Z, a-z) are allowed", i+2, row[1]))
		}

		genderLowered := strings.ToLower(row[3])
		if genderLowered != "male" && genderLowered != "female" {
			errList = append(errList, fmt.Sprintf("row %d: Invalid student gender: %s, gender should be male / female", i+2, genderLowered))
		}

		_, err = strconv.Atoi(row[4])
		if err != nil {
			errList = append(errList, fmt.Sprintf("row %d: Invalid student telephone: %s, telephone should be numeric", i+2, row[4]))
		}

		if len(row[4]) > 15 {
			errList = append(errList, fmt.Sprintf("row %d: Invalid student telephone: %s, telephone should be 15 number max", i+2, genderLowered))
		}

		studentDataHolder = domain.Student{
			Name:       row[0],
			Grade:      convertedGrade,
			GradeLabel: row[2],
			Gender:     genderLowered,
			Telephone:  row[4],
			ParentID:   0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		genderParentLowered := strings.ToLower(row[6])
		if genderParentLowered != "male" && genderParentLowered != "female" {
			errList = append(errList, fmt.Sprintf("row %d: Invalid gender: %s, gender should be male / female", i+2, genderParentLowered))
		}

		_, err = strconv.Atoi(row[7])
		if err != nil {
			errList = append(errList, fmt.Sprintf("row %d: Invalid parent telephone: %s, telephone should be numeric", i+2, row[7]))
		}

		if len(row[7]) > 15 {
			errList = append(errList, fmt.Sprintf("row %d: Invalid parent telephone: %s, telephone should be 15 number max", i+2, genderParentLowered))
		}

		parentDataHolder = domain.Parent{
			Name:      row[5],
			Gender:    row[6],
			Telephone: row[7],
			Email:     getStringPointer(row[8]),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		studNParent := domain.StudentAndParent{
			Student: studentDataHolder,
			Parent:  parentDataHolder,
		}

		listStudentAndParent = append(listStudentAndParent, studNParent)
	}

	if len(errList) > 0 {
		return &errList, nil, nil
	}

	duplicates, _ := sph.uc.ImportCSV(c, &listStudentAndParent)

	if duplicates != nil && len(*duplicates) > 0 {
		return nil, duplicates, nil
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
	userToken, _ := c.Locals("user").(*domain.Claims)

	id := c.Params("id")
	if id == "" {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID Required",
		})
	}

	convertetID, err := strconv.Atoi(id)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid ID",
			"error":   err.Error(),
		})
	}

	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid request body",
		})
	}

	_, err = govalidator.ValidateStruct(&req)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		validationErrors := govalidator.ErrorsByField(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"errors":  validationErrors,
			"message": "Invalid request body",
		})
	}

	errList := sph.uc.UpdateStudentAndParent(c.Context(), convertetID, &req)
	if errList != nil && len(*errList) > 0 {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "UpdateStudentAndParent")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   *errList,
			"message": "Failed to update student and parent",
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "UpdateStudentAndParent")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent updated successfully",
	})
}

func (sph *studentParentHandler) DeleteStudentAndParent(c *fiber.Ctx) error {
	userToken, _ := c.Locals("user").(*domain.Claims)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DeleteStudentAndParent")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	if err := sph.uc.DeleteStudentAndParent(c.Context(), id); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DeleteStudentAndParent")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete student",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DeleteStudentAndParent")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student deleted successfully",
	})
}

func (sph *studentParentHandler) GetStudentDetailsByID(c *fiber.Ctx) error {
	userToken, _ := c.Locals("user").(*domain.Claims)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "GetStudentDetailsByID")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	student, err := sph.uc.GetStudentDetailsByID(c.Context(), id)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetStudentDetailsByID")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get student",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetStudentDetailsByID")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent retrieved successfully",
		"data":    student.Student,
	})
}

func (sph *studentParentHandler) DataChangeRequest(c *fiber.Ctx) error {
	guess := "Guest"
	var datas domain.DataChangeRequest

	err := c.BodyParser(&datas)
	if err != nil {
		config.PrintLogInfo(&guess, fiber.StatusBadRequest, "DataChangeRequest")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Request",
			"error":   err,
		})
	}

	err = sph.uc.DataChangeRequest(c.Context(), datas)
	if err != nil {
		config.PrintLogInfo(&guess, fiber.StatusInternalServerError, "DataChangeRequest")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to send data change request",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&guess, fiber.StatusOK, "DataChangeRequest")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully sent data changes",
	})
}
