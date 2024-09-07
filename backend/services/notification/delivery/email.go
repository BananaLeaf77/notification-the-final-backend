package delivery

import (
	"notification/domain"

	"github.com/gofiber/fiber/v2"
)

type emailSenderHandler struct {
	suc domain.EmailSMTPUseCase
}

// NewEmailerDelivery sets up the route for the email sender handler
func NewEmailerDelivery(app *fiber.App, uc domain.EmailSMTPUseCase) {
	handler := &emailSenderHandler{
		suc: uc,
	}

	route := app.Group("/email")
	route.Post("/send-mass", handler.sendMassHandler)
}

func (h *emailSenderHandler) sendMassHandler(c *fiber.Ctx) error {

	wg.Add(1)
	defer wg.Done()

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
			"error":  "failed to send emails",
			"detail": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "emails sent successfully",
	})
}
