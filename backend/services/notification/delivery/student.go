package delivery

import (
	"notification/domain"

	"github.com/gofiber/fiber/v2"
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

func (sh *studentHandler) deliveryGetAllStudent(c *fiber.Ctx) error {
	students, err := sh.suc.GetAllStudentUC(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get all students",
			"error":   err.Error(),
		})
	}

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
