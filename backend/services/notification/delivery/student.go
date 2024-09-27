package delivery

import (
	"notification/config"
	"notification/domain"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
)

var wg *sync.WaitGroup = config.GetWaitGroupInstance()

type studentHandler struct {
	suc domain.StudentUseCase
}

func NewStudentDelivery(app *fiber.App, uc domain.StudentUseCase) {
	handler := &studentHandler{
		suc: uc,
	}

	route := app.Group("/student")
	route.Post("/insert", handler.deliveryInsertStudent)
	route.Get("/get_all", handler.deliveryGetAllStudent)
	route.Get("/get/:id", handler.deliveryGetStudentByID)
	route.Put("/modify/:id", handler.deliveryModifyStudent)
	route.Delete("/rm/:id", handler.deliveryDeleteStudent)
	route.Get("/download_input_template", handler.deliveryDownloadTemplate)
}

func (sh *studentHandler) deliveryInsertStudent(c *fiber.Ctx) error {
	wg.Add(1)
	defer wg.Done()
	var student domain.Student
	if err := c.BodyParser(&student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if err := sh.suc.CreateStudentUC(c.Context(), &student); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create student",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Student created successfully",
		"data":    student,
	})
}

func (sh *studentHandler) deliveryGetAllStudent(c *fiber.Ctx) error {
	students, err := sh.suc.GetAllStudentUC(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get all students",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Students retrieved successfully",
		"data":    students,
	})
}

func (sh *studentHandler) deliveryGetStudentByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	student, err := sh.suc.GetStudentByIDUC(c.Context(), id)
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
		"data":    student,
	})
}

func (sh *studentHandler) deliveryModifyStudent(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	var student domain.Student
	if err := c.BodyParser(&student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}
	student.ID = id

	if err := sh.suc.UpdateStudentUC(c.Context(), &student); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update student",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student updated successfully",
		"data":    student,
	})
}

func (sh *studentHandler) deliveryDeleteStudent(c *fiber.Ctx) error {
	wg.Add(1)
	defer wg.Done()

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	if err := sh.suc.DeleteStudentUC(c.Context(), id); err != nil {
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

func (sh *studentHandler) deliveryDownloadTemplate(c *fiber.Ctx) error {

	filePath, err := sh.suc.DownloadInputDataTemplate(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get the input data template",
			"error":   err.Error(),
		})
	}

	c.Set("Content-Disposition", "attachment; filename=input_data_template.csv")
	c.Set("Content-Type", "text/csv")

	return c.SendFile(*filePath)
}
