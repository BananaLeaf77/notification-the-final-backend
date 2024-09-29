package delivery

import (
	"context"
	"notification/domain"
	"notification/middleware"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	uc domain.UserUseCase
}

func NewUserHandler(app *fiber.App, useCase domain.UserUseCase) {
	handler := &UserHandler{
		uc: useCase,
	}

	app.Post("/login", handler.Login)
	app.Post("/create-staff", handler.CreateStaff)
}

func (uh *UserHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"success": false,
		})
	}

	user, err := uh.uc.FindUserByUsername(context.Background(), req.Username)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Invalid username or password",
			"success": false,
		})
	}

	if req.Username != "admin" {
		// Compare hashed password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Invalid username or password",
				"success": false,
			})
		}
	}

	// Generate JWT
	token, err := middleware.GenerateJWT(user.Username, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Could not generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token": token,
	})
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

