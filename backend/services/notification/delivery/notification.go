package delivery

import (
	"notification/domain"

	"github.com/gofiber/fiber/v2"
)

type notifHandler struct {
	uc domain.NotificationUseCase
}

func NewNotificationHandler(app *fiber.App, uc domain.NotificationUseCase) {
	handler := &notifHandler{
		uc: uc,
	}

	group := app.Group("/notification")
	group.Get("/truancy-history", handler.GetAllAttendanceNotificationHistory)
}

func (nh *notifHandler) GetAllAttendanceNotificationHistory(c *fiber.Ctx) error {
	datas, err := nh.uc.GetAllAttendanceNotificationHistory(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get all truancy history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully retrieved all truancy history",
		"data":    datas,
	})
}
