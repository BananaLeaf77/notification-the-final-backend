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

	route.Post("/insert", handler.deliveryInsertStudent)
	route.Get("/get_all", handler.deliveryGetAllStudent)
	route.Get("/get/:id", handler.deliveryGetStudentByID)
	route.Put("/modify/:id", handler.deliveryModifyStudent)
	route.Delete("/rm/:id", handler.deliveryDeleteStudent)

}
