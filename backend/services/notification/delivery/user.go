package delivery

import (
	"notification/domain"
	"notification/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	uc domain.UserUseCase
}

func NewUserHandler(app *fiber.App, useCase domain.UserUseCase) {
	handler := &UserHandler{
		uc: useCase,
	}

	app.Post("/staff/create-staff", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateStaff)
	app.Get("/staff/get-all", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllStaff)
	app.Delete("/staff/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStaff)
	app.Get("/staff/details/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetStaffDetail)
	app.Put("/staff/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.ModifyStaff)
}

func (uh *UserHandler) CreateStaff(c *fiber.Ctx) error {
	var payload domain.User
	payload.Role = "staff"

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	_, err := uh.uc.CreateStaff(c.Context(), &payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Account created successfully",
	})
}

func (uh *UserHandler) GetAllStaff(c *fiber.Ctx) error {
	v, err := uh.uc.GetAllStaff(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff retrieved successfully",
		"data":    v,
	})
}

func (uh *UserHandler) DeleteStaff(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
		})
	}

	err = uh.uc.DeleteStaff(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff deleted successfully",
	})
}

func (uh *UserHandler) GetStaffDetail(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
		})
	}

	v, err := uh.uc.GetStaffDetail(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff retrieved successfully",
		"data":    v,
	})

}

func (uh *UserHandler) ModifyStaff(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid staff ID",
		})
	}

	var payload domain.User
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	payload.Role = "staff"

	err = uh.uc.UpdateStaff(c.Context(), id, &payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to modify staff",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff modified successfully",
	})
}
