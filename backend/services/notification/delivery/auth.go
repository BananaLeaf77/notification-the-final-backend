package delivery

import (
	"notification/domain"

	"github.com/gofiber/fiber/v2"
)

type userHandler struct {
	uc domain.AuthUseCase
}

func NewUserAuthHandler(app *fiber.App, uc domain.AuthUseCase) {
	handler := &userHandler{
		uc: uc,
	}

	app.Post("/login", handler.Login)
}

func (h *userHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	dataList, err := h.uc.Login(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to login",
		})
	}

	theRole := (*dataList)[0]
	theToken := (*dataList)[1]

	return c.Status(fiber.StatusOK).JSON(domain.LoginResponse{
		Token: theToken,
		Role:  theRole,
	})
}
