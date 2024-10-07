package delivery

import (
	"notification/domain"

	"github.com/gofiber/fiber/v2"
)

type senderHandler struct {
	suc domain.SenderUseCase
}

// NewEmailerDelivery sets up the route for the email sender handler
func NewSenderDelivery(app *fiber.App, uc domain.SenderUseCase) {
	handler := &senderHandler{
		suc: uc,
	}

	route := app.Group("/sender")
	route.Post("/send-mass", handler.sendMassHandler)
}

func (h *senderHandler) sendMassHandler(c *fiber.Ctx) error {
	var payload struct {
		IDs []int `json:"ids"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Call the use case with the parsed ID list.
	if err := h.suc.SendMass(c.Context(), &payload.IDs); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "failed to send notifications",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "notifications sent successfully",
	})
}
