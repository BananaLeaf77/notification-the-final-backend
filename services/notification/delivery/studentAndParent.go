package delivery

import (
	"context"
	"encoding/csv"
	"errors"
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
	"unicode"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
)

type studentParentHandler struct {
	uc domain.StudentParentUseCase
}

func NewStudentParentHandlerDeploy(app *fiber.App, useCase domain.StudentParentUseCase) {
	handler := &studentParentHandler{
		uc: useCase,
	}

	route := app.Group("/student-and-parent")
	route.Post("/insert", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateStudentAndParent)
	route.Post("/import", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.UploadAndImport)
	route.Put("/modify/:student_nsn", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.UpdateStudentAndParent)
	// route.Delete("/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStudentAndParent)
	route.Get("/student/:student_nsn", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.GetStudentDetailsByID)
	route.Post("/req/data-change-request", middleware.AuthRequired(), middleware.RoleRequired("staff"), handler.DataChangeRequest)
	route.Get("/get-all-data-change-request", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllDataChangeRequest)
	route.Get("/get-all-data-change-request/:request_id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllDataChangeRequestByID)
	// route.Post("/rms", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.SPMassDelete)
	route.Get("/download-template", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DownloadTemplate)
	route.Delete("/review/dcr/:request_id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteDCR)
	route.Post("/approve/dcr", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.ApproveDCR)
}

func (sph *studentParentHandler) ApproveDCR(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	var payloadReadyForApprove struct {
		Name         *string    `json:"name"`
		Gender       *string    `json:"gender"`
		Telephone    *string    `json:"telephone"`
		OldTelephone *string    `json:"old_telephone"`
		Email        *string    `json:"email"`
		UpdatedAt    *time.Time `json:"updated_at"`
	}

	if err := c.BodyParser(&payloadReadyForApprove); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "ApproveDCR")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Invalid data",
		})
	}

	if payloadReadyForApprove.OldTelephone == nil || *payloadReadyForApprove.OldTelephone == "" {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "ApproveDCR")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Old telephone is required",
		})
	}

	repoPayload := map[string]interface{}{
		"name":         payloadReadyForApprove.Name,
		"gender":       payloadReadyForApprove.Gender,
		"telephone":    payloadReadyForApprove.Telephone,
		"oldTelephone": payloadReadyForApprove.OldTelephone,
		"email":        payloadReadyForApprove.Email,
	}

	allocated, err := sph.uc.ApproveDCR(c.Context(), repoPayload)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "ApproveDCR")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to approve data changes",
		})
	}

	if allocated != nil {
		msgs := fmt.Sprintf("Data changes approved, %s", *allocated)
		config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "ApproveDCR")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": msgs,
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "ApproveDCR")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Data changes approved",
	})
}

func (sph *studentParentHandler) DownloadTemplate(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	filePath := "./template/sinoan_template.csv"

	c.Set(fiber.HeaderContentDisposition, `attachment; filename="sinoan_template.csv"`)
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

func (sph *studentParentHandler) DeleteDCR(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	id := c.Params("request_id")
	convertedID, err := strconv.Atoi(id)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DeleteDCR")
		return c.Status(fiber.StatusBadRequest).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Converter Failure on request_id",
		}))
	}
	err = sph.uc.DeleteDCR(c.Context(), convertedID)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DeleteDCR")
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to delete data change request",
		}))
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DeleteDCR")
	return c.Status(fiber.StatusOK).JSON((fiber.Map{
		"success": true,
		"message": "Data Change Request deleted successfully",
	}))
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

func (sph *studentParentHandler) GetAllDataChangeRequestByID(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	id := c.Params("request_id")
	v, err := strconv.Atoi(id)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "GetAllDataChangeRequestByID")
		return c.Status(fiber.StatusBadRequest).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Converter failure",
		}))
	}

	data, err := sph.uc.GetAllDataChangeRequestByID(c.Context(), v)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetAllDataChangeRequestByID")
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to get data change request",
		}))
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetAllDataChangeRequestByID")
	return c.Status(fiber.StatusOK).JSON((fiber.Map{
		"success": true,
		"message": "Data Change Request Retrieved Successfully",
		"data":    data,
	}))

}

// func (sph *studentParentHandler) SPMassDelete(c *fiber.Ctx) error {
// 	userToken := c.Locals("user").(*domain.Claims)
// 	var payload struct {
// 		IDS []int `json:"student_ids"`
// 	}

// 	err := c.BodyParser(&payload)
// 	if err != nil {
// 		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "SPMassDelete")
// 		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
// 			"success": false,
// 			"error":   err.Error(),
// 			"message": "Failed to delete students",
// 		}))
// 	}

// 	err = sph.uc.SPMassDelete(c.Context(), &payload.IDS)
// 	if err != nil {
// 		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "SPMassDelete")
// 		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
// 			"success": false,
// 			"error":   err.Error(),
// 			"message": "Failed to delete students",
// 		}))
// 	}

// 	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "SPMassDelete")
// 	return c.Status(fiber.StatusOK).JSON((fiber.Map{
// 		"success": true,
// 		"message": "Students deleted successfully",
// 	}))
// }

func (sph *studentParentHandler) CreateStudentAndParent(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "CreateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   []string{"Invalid request body: %v", err.Error()},
			"message": "Invalid request body",
		})
	}

	if req.Student.GradeLabel == "" {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "CreateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   []string{"Invalid Grade Label: Grade Label is required"},
			"message": "Invalid Student request body",
		})
	}

	allocated, errList := sph.uc.CreateStudentAndParentUC(c.Context(), &req)
	if errList != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "CreateStudentAndParent")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   errList,
			"message": "Failed to Create Student and Parent",
		})
	}

	if allocated != nil {
		msgs := fmt.Sprintf("Student and Parent created successfully, %s", *allocated)
		config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "ApproveDCR")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": msgs,
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Import Failure, bad input found.",
			"error":   badRequests,
		})
	}

	if internalServerResponse != nil && len(*internalServerResponse) > 0 {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "UploadAndImport")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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
	var duplicateErrList []string
	var listStudentAndParent []domain.StudentAndParent

	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open CSV file: %v", err)
	}

	defer func() {
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete file: %v", err)
		}
	}()

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV file: %v", err)
	}

	// Precompile regular expressions
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Skip the header row
	for i, row := range records[1:] {
		if len(row) < 10 {
			errList = append(errList, fmt.Sprintf("row %d: insufficient columns, expected 10 columns", i+2))
			continue
		}

		// Validate Student Data
		trimmedStudentNSN := strings.TrimSpace(row[0])
		row[0] = trimmedStudentNSN

		trimmedStudentGrade := strings.TrimSpace(row[2])
		row[2] = trimmedStudentGrade

		trimmedStudentGradeLabel := strings.TrimSpace(row[3])
		row[3] = trimmedStudentGradeLabel

		trimmedStudentTelephone := strings.TrimSpace(row[5])
		row[5] = trimmedStudentTelephone

		studentErrors := validateStudent(row[:6], i+2)
		if len(studentErrors) > 0 {
			errList = append(errList, studentErrors...)
		}

		// Validate Parent Data
		trimmedParentTelephone := strings.TrimSpace(row[8])
		row[8] = trimmedParentTelephone

		trimmedEmail := strings.TrimSpace(row[9])
		row[9] = trimmedEmail
		parentErrors := validateParent(row[6:], i+2, emailRegex)
		if len(parentErrors) > 0 {
			errList = append(errList, parentErrors...)
		}

		// Populate student and parent data if no errors
		if len(studentErrors) == 0 && len(parentErrors) == 0 {
			student := domain.Student{
				StudentNSN: row[0],
				Name:       row[1],
				Grade:      mustAtoi(row[2]),
				GradeLabel: strings.ToUpper(row[3]),
				Gender:     strings.ToLower(row[4]),
				Telephone:  row[5],
				ParentID:   0,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			parent := domain.Parent{
				Name:      row[6],
				Gender:    strings.ToLower(row[7]),
				Telephone: row[8],
				Email:     getStringPointer(row[9]),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			listStudentAndParent = append(listStudentAndParent, domain.StudentAndParent{
				Student: student,
				Parent:  parent,
			})
		}
	}

	// Check for duplicates in listStudentAndParent
	duplicateErrList = checkDuplicates(listStudentAndParent)

	// Return errors if any
	if len(errList) > 0 || len(duplicateErrList) > 0 {
		combinedErrList := append(errList, duplicateErrList...)
		return &combinedErrList, nil, nil
	}

	// Import data if no errors
	duplicates, _ := sph.uc.ImportCSV(c, &listStudentAndParent)
	if duplicates != nil && len(*duplicates) > 0 {
		return nil, duplicates, nil
	}

	return nil, nil, nil
}

// Helper function to validate student data
func validateStudent(row []string, rowNum int) []string {
	var errList []string

	// Validate NSN
	if row[0] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Student NSN cannot be empty", rowNum))
	} else if len(row[0]) > 10 {
		errList = append(errList, fmt.Sprintf("row %d: Student NSN cannot be more than 10 characters", rowNum))
	} else if !isNumeric(row[0]) {
		errList = append(errList, fmt.Sprintf("row %d: Student NSN must contain only digits", rowNum))
	}

	// Validate Student Name
	if row[1] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Student name cannot be empty", rowNum))
	} else if len(row[1]) > 150 {
		errList = append(errList, fmt.Sprintf("row %d: Student name cannot be more than 150 characters", rowNum))
	} else if containsDigit(row[1]) {
		errList = append(errList, fmt.Sprintf("row %d: Student name cannot contain digits", rowNum))
	}

	// Validate Grade
	if row[2] == "" {
		errList = append(errList, fmt.Sprintf("row %d: grade cannot be empty", rowNum))
	} else if grade, err := strconv.Atoi(row[2]); err != nil || grade > 99 {
		errList = append(errList, fmt.Sprintf("row %d: Student grade: %s, must be a number and less than 100", rowNum, row[2]))
	}

	// Validate Grade Label
	if row[3] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Student grade label cannot be empty", rowNum))
	} else if len(row[3]) > 5 {
		errList = append(errList, fmt.Sprintf("row %d: Student grade label cannot be more than 5 characters", rowNum))
	}
	// Validate Gender
	if row[4] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Student gender cannot be empty", rowNum))
	} else if gender := strings.ToLower(row[4]); gender != "male" && gender != "female" {
		errList = append(errList, fmt.Sprintf("row %d: invalid gender: %s, must be 'male' or 'female'", rowNum, row[4]))
	}

	// Validate Telephone
	if row[5] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Student telephone cannot be empty", rowNum))
	} else if len(row[5]) > 13 {
		errList = append(errList, fmt.Sprintf("row %d: Student telephone cannot be more than 13 characters", rowNum))
	} else if !isNumeric(row[5]) {
		errList = append(errList, fmt.Sprintf("row %d: Student telephone must contain only digits", rowNum))
	}

	return errList
}

// Helper function to validate parent data
func validateParent(row []string, rowNum int, emailRegex *regexp.Regexp) []string {
	var errList []string

	// Validate Parent Name
	if row[0] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Parent name cannot be empty", rowNum))
	} else if len(row[0]) > 150 {
		errList = append(errList, fmt.Sprintf("row %d: Parent name cannot be more than 150 characters", rowNum))
	} else if containsDigit(row[0]) {
		errList = append(errList, fmt.Sprintf("row %d: Parent name cannot contain digits", rowNum))
	}

	// Validate Gender
	if row[1] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Parent gender cannot be empty", rowNum))
	} else if gender := strings.ToLower(row[1]); gender != "male" && gender != "female" {
		errList = append(errList, fmt.Sprintf("row %d: Parent gender: %s, must be 'male' or 'female'", rowNum, row[1]))
	}

	// Validate Telephone
	if row[2] == "" {
		errList = append(errList, fmt.Sprintf("row %d: Parent telephone cannot be empty", rowNum))
	} else if len(row[2]) > 13 {
		errList = append(errList, fmt.Sprintf("row %d: Parent telephone cannot be more than 13 characters", rowNum))
	} else if !isNumeric(row[2]) {
		errList = append(errList, fmt.Sprintf("row %d: Parent telephone must contain only digits", rowNum))
	}

	// Validate Email (optional)
	if row[3] != "" {
		if len(row[3]) > 255 {
			errList = append(errList, fmt.Sprintf("row %d: Parent email cannot be more than 255 characters", rowNum))
		} else if !emailRegex.MatchString(row[3]) {
			errList = append(errList, fmt.Sprintf("row %d: Parent email format is invalid: %s", rowNum, row[3]))
		}
	}

	return errList
}

func checkDuplicates(list []domain.StudentAndParent) []string {
	var duplicateErrList []string
	seenNames := make(map[string]int)             // Track seen student names
	seenStudentTelephones := make(map[string]int) // Track seen student telephones
	seenNSNs := make(map[string]int)              // Track seen NSNs

	for i, item := range list {
		// Check for duplicate NSNs
		if j, exists := seenNSNs[item.Student.StudentNSN]; exists {
			duplicateErrList = append(duplicateErrList, fmt.Sprintf("duplicate student NSN: %s found in rows %d and %d", item.Student.StudentNSN, j+2, i+2))
		} else {
			seenNSNs[item.Student.StudentNSN] = i
		}

		// Check for duplicate student names
		if j, exists := seenNames[item.Student.Name]; exists {
			duplicateErrList = append(duplicateErrList, fmt.Sprintf("duplicate student name: %s found in rows %d and %d", item.Student.Name, j+2, i+2))
		} else {
			seenNames[item.Student.Name] = i
		}

		// Check for duplicate student telephones
		if j, exists := seenStudentTelephones[item.Student.Telephone]; exists {
			duplicateErrList = append(duplicateErrList, fmt.Sprintf("duplicate student telephone: %s found in rows %d and %d", item.Student.Telephone, j+2, i+2))
		} else {
			seenStudentTelephones[item.Student.Telephone] = i
		}

		// Check if student and parent have the same telephone number
		if item.Student.Telephone == item.Parent.Telephone {
			duplicateErrList = append(duplicateErrList, fmt.Sprintf("row %d: student and parent have the same telephone number: %s", i+2, item.Student.Telephone))
		}
	}

	return duplicateErrList
}

// Helper function to check if a string contains only digits
func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// Helper function to check if a string contains any digits
func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// Helper function to convert string to integer (panics on error)
func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("invalid integer conversion: %s", s))
	}
	return i
}

// Helper function to get a string pointer
func getStringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (sph *studentParentHandler) UpdateStudentAndParent(c *fiber.Ctx) error {
	userToken, _ := c.Locals("user").(*domain.Claims)

	studentNSN := c.Params("student_nsn")
	if studentNSN == "" {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "student nsn Required",
		})
	}

	var req domain.StudentAndParent
	if err := c.BodyParser(&req); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   []string{"Invalid request body: %v", err.Error()},
			"message": "Invalid request body",
		})
	}

	if req.Student.StudentNSN == "" || req.Student.Name == "" || req.Student.Grade == 0 || req.Student.GradeLabel == "" || req.Student.Gender == "" || req.Student.Telephone == "" || req.Parent.Name == "" || req.Parent.Telephone == "" || req.Parent.Gender == "" {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   []string{"fields cannot be blank"}, // Error as an array
			"message": "Invalid request body",
		})
	}

	var validatorResponse []string
	_, err := govalidator.ValidateStruct(&req)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "UpdateStudentAndParent")
		validationErrors := govalidator.ErrorsByField(err)
		for i := range validationErrors {
			validatorResponse = append(validatorResponse, validationErrors[i])
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   validatorResponse,
			"message": "Invalid request body",
		})
	}

	allocated, errList := sph.uc.UpdateStudentAndParent(c.Context(), studentNSN, &req)
	if errList != nil && len(*errList) > 0 {
		fmt.Println(errList)
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "UpdateStudentAndParent")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   *errList,
			"message": "Failed to update student and parent",
		})
	}

	if allocated != nil {
		msgs := fmt.Sprintf("Student and Parent updated successfully, %s", *allocated)
		config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "ApproveDCR")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": msgs,
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "UpdateStudentAndParent")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student and Parent updated successfully",
	})
}

// func (sph *studentParentHandler) DeleteStudentAndParent(c *fiber.Ctx) error {
// 	userToken, _ := c.Locals("user").(*domain.Claims)
// 	id, err := strconv.Atoi(c.Params("id"))
// 	if err != nil {
// 		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DeleteStudentAndParent")

// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Invalid student ID",
// 			"error":   err.Error(),
// 		})
// 	}

// 	if err := sph.uc.DeleteStudentAndParent(c.Context(), id); err != nil {
// 		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DeleteStudentAndParent")

// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"success": false,
// 			"message": "Failed to delete student",
// 			"error":   err.Error(),
// 		})
// 	}

// 	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DeleteStudentAndParent")
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"success": true,
// 		"message": "Student deleted successfully",
// 	})
// }

func (sph *studentParentHandler) GetStudentDetailsByID(c *fiber.Ctx) error {
	userToken, _ := c.Locals("user").(*domain.Claims)

	studentNSN := c.Params("student_nsn")
	student, err := sph.uc.GetStudentDetailsByID(c.Context(), studentNSN)
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

func ValidateDataChangeRequest(data *domain.ParentDataChangeRequest) error {
	if data == nil {
		return errors.New("request cannot be nil")
	}

	if (data.NewParentName == nil || *data.NewParentName == "") &&
		(data.NewParentTelephone == nil || *data.NewParentTelephone == "") &&
		(data.NewParentEmail == nil || *data.NewParentEmail == "") &&
		(data.NewParentGender == nil || *data.NewParentGender == "") {
		return errors.New("please input at least one new data field")
	}

	return nil
}

func (sph *studentParentHandler) DataChangeRequest(c *fiber.Ctx) error {
	var datas domain.ParentDataChangeRequest
	userToken, _ := c.Locals("user").(*domain.Claims)
	err := c.BodyParser(&datas)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DataChangeRequest")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Request",
			"error":   err,
		})
	}

	if datas.NewParentTelephone != nil && *datas.NewParentTelephone == datas.OldParentTelephone {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DataChangeRequest")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Request",
			"error":   "new parent telephone should not have the same value as old parent telephone",
		})
	}

	err = ValidateDataChangeRequest(&datas)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DataChangeRequest")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid Request",
			"error":   err.Error(),
		})
	}

	err = sph.uc.DataChangeRequest(c.Context(), datas, userToken.UserID)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DataChangeRequest")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to send data change request",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DataChangeRequest")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully sent data changes request",
	})
}
