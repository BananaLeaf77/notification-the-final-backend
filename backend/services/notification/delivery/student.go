package delivery

import (
	"fmt"
	"notification/domain"
	"notification/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type studentHandler struct {
	suc domain.StudentUseCase
}

func NewStudentDelivery(app *fiber.App, uc domain.StudentUseCase) {
	handler := &studentHandler{
		suc: uc,
	}

	route := app.Group("/student")
	route.Get("/get-all", handler.deliveryGetAllStudent)
	route.Get("/download_input_template", handler.deliveryDownloadTemplate)
}

func NewStudentDeliveryDeploy(app *fiber.App, uc domain.StudentUseCase) {
	handler := &studentHandler{
		suc: uc,
	}

	route := app.Group("/student")
	route.Get("/get-all", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.deliveryGetAllStudent)
	route.Get("/download_input_template", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.deliveryDownloadTemplate)
}

func (sh *studentHandler) deliveryGetAllStudent(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*domain.Claims)
	if !ok {
		log.Error("[Error] Failed to extract user token from context")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	log.Info(fmt.Sprintf("User: %s => Request to get all students", userToken.Username))

	students, err := sh.suc.GetAllStudentUC(c.Context())
	if err != nil {
		log.Error(fmt.Sprintf("User: %s => Failed to get all students: %v", userToken.Username, err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve students",
			"error":   err.Error(),
		})
	}

	log.Info(fmt.Sprintf("User: %s => Successfully retrieved all students", userToken.Username))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Students retrieved successfully",
		"data":    students,
	})
}

func (sh *studentHandler) deliveryDownloadTemplate(c *fiber.Ctx) error {

	filePath, err := sh.suc.DownloadInputDataTemplate(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get the input data template",
			"error":   err.Error(),
		})
	}

	c.Set("Content-Disposition", "attachment; filename=input_data_template.csv")
	c.Set("Content-Type", "text/csv")

	return c.SendFile(*filePath)
}
