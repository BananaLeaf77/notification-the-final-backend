package delivery

import (
	"notification/config"
	"notification/domain"
	"notification/middleware"

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

func NewNotificationHandlerDeploy(app *fiber.App, uc domain.NotificationUseCase) {
	handler := &notifHandler{
		uc: uc,
	}

	group := app.Group("/notification")
	group.Get("/truancy-history", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllAttendanceNotificationHistory)
}

func (nh *notifHandler) GetAllAttendanceNotificationHistory(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	datas, err := nh.uc.GetAllAttendanceNotificationHistory(c.Context())
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetAllAttendanceNotificationHistory")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get all truancy history",
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetAllAttendanceNotificationHistory")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Successfully retrieved all truancy history",
		"data":    datas,
	})
}
