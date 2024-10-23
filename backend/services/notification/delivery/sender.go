package delivery

import (
	"notification/domain"
	"notification/middleware"

	"github.com/gofiber/fiber/v2"
)

type senderHandler struct {
	suc domain.SenderUseCase
}

func NewSenderDelivery(app *fiber.App, uc domain.SenderUseCase) {
	handler := &senderHandler{
		suc: uc,
	}

	route := app.Group("/sender")
	route.Post("/send-mass", handler.sendMassHandler)
}

func NewSenderDeliveryDeploy(app *fiber.App, uc domain.SenderUseCase) {
	handler := &senderHandler{
		suc: uc,
	}

	route := app.Group("/sender")
	route.Post("/send-mass", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.sendMassHandler)
}

func (h *senderHandler) sendMassHandler(c *fiber.Ctx) error {
	var payload struct {
		IDs []int `json:"ids"`
	}

	userToken := c.Locals("user").(*domain.Claims)
	userID := userToken.UserID

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.suc.SendMass(c.Context(), &payload.IDs, &userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to send notifications",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "notifications sent successfully",
	})
}
