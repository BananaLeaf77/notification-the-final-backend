package delivery

import (
	"context"
	"notification/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

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
}

func (sh *studentHandler) deliveryInsertStudent(c *fiber.Ctx) error {
	var student domain.Student
	if err := c.BodyParser(&student); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	ctx := context.Background()
	if err := sh.suc.CreateStudentUC(ctx, &student); err != nil {
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
	ctx := context.Background()
	students, err := sh.suc.GetAllStudentUC(ctx)
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

	ctx := context.Background()
	student, err := sh.suc.GetStudentByIDUC(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get student",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Student retrieved successfully",
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

	ctx := context.Background()
	if err := sh.suc.UpdateStudentUC(ctx, &student); err != nil {
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
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid student ID",
			"error":   err.Error(),
		})
	}

	ctx := context.Background()
	if err := sh.suc.DeleteStudentUC(ctx, id); err != nil {
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
