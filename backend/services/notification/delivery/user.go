package delivery

import (
	"notification/domain"
	"notification/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type uHandler struct {
	uc domain.UserUseCase
}

func NewUserHandler(app *fiber.App, useCase domain.UserUseCase) {
	handler := &uHandler{
		uc: useCase,
	}

	app.Post("/staff/create-staff", handler.CreateStaff)
	app.Get("/staff/get-all", handler.GetAllStaff)
	app.Delete("/staff/rm/:id", handler.DeleteStaff)
	app.Get("/staff/details/:id", handler.GetStaffDetail)
	app.Put("/staff/modify/:id", handler.ModifyStaff)
	app.Post("/create-class", handler.CreateClass)
}

func NewUserHandlerDeploy(app *fiber.App, useCase domain.UserUseCase) {
	handler := &uHandler{
		uc: useCase,
	}

	app.Post("/staff/create-staff", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateStaff)
	app.Get("/staff/get-all", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllStaff)
	app.Delete("/staff/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStaff)
	app.Get("/staff/details/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetStaffDetail)
	app.Put("/staff/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.ModifyStaff)
}

func (uh *uHandler) CreateClass(c *fiber.Ctx) error {
	var class domain.Class
	err := c.BodyParser(&class)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	err = uh.uc.CreateClass(c.Context(), &class)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Class created successfully",
	})

}

func (uh *uHandler) CreateStaff(c *fiber.Ctx) error {
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

func (uh *uHandler) GetAllStaff(c *fiber.Ctx) error {
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

func (uh *uHandler) DeleteStaff(c *fiber.Ctx) error {
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

func (uh *uHandler) GetStaffDetail(c *fiber.Ctx) error {
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

func (uh *uHandler) ModifyStaff(c *fiber.Ctx) error {
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
