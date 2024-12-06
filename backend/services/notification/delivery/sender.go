package delivery

import (
	"fmt"
	"notification/config"
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
	route.Post("/send-mass/exam-result", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.SendTestScores)
}

func (h *senderHandler) SendTestScores(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	var payload struct {
		ExamType string `json:"exam_type"`
	}

	err := c.BodyParser(&payload)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "SendTestScores")
		return c.Status(fiber.StatusBadRequest).JSON((fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to announce test scores",
		}))
	}

	fmt.Println(payload.ExamType)

	err = h.suc.SendTestScores(c.Context(), payload.ExamType)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "SendTestScores")
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to announce test scores",
		}))
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "SendTestScores")
	return c.Status(fiber.StatusOK).JSON((fiber.Map{
		"success": true,
		"message": "Successfully announce test scores",
	}))
}

func (h *senderHandler) sendMassHandler(c *fiber.Ctx) error {
	var payload struct {
		IDs       []int `json:"ids"`
		SubjectID int   `json:"subject_id"`
	}

	userToken := c.Locals("user").(*domain.Claims)
	userID := userToken.UserID

	if err := c.BodyParser(&payload); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "sendMassHandler")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"success": false,
			"message": "Failed to announce attendance",
		})
	}

	if err := h.suc.SendMass(c.Context(), &payload.IDs, &userID, payload.SubjectID); err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "sendMassHandler")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to send notifications",
			"detail":  err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "sendMassHandler")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "notifications sent successfully",
		"success": true,
	})
}
