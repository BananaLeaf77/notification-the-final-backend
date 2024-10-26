package delivery

import (
	"notification/domain"
	"notification/middleware"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type uHandler struct {
	uc domain.UserUseCase
}

func NewUserHandler(app *fiber.App, useCase domain.UserUseCase) {
	handler := &uHandler{
		uc: useCase,
	}
	group := app.Group("/user")
	group.Post("/create-staff", handler.CreateStaff)
	group.Get("/get-all", handler.GetAllStaff)
	group.Delete("/rm/:id", handler.DeleteStaff)
	group.Get("/details/:id", handler.GetStaffDetail)
	group.Put("/modify/:id", handler.ModifyStaff)

	group.Post("/add-subject", handler.CreateSubject)
	group.Post("/add-subject-bulk", handler.CreateSubjectBulk)
	group.Get("/subject/all", handler.GetAllSubject)
	group.Put("/subject/modify/:id", handler.UpdateSubject)
	group.Delete("/subject/rm/:id", handler.DeleteSubject)
}

func NewUserHandlerDeploy(app *fiber.App, useCase domain.UserUseCase) {
	handler := &uHandler{
		uc: useCase,
	}
	group := app.Group("/user")
	app.Post("/create-staff", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateStaff)
	app.Get("/get-all", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllStaff)
	app.Delete("/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStaff)
	app.Get("/details/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetStaffDetail)
	app.Put("/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.ModifyStaff)

	group.Post("/add-subject", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateSubject)
	group.Post("/add-subject-bulk", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateSubjectBulk)
	group.Get("/subject/all", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllSubject)
	group.Put("/subject/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.UpdateSubject)
	group.Delete("/subject/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteSubject)
}

func (uh *uHandler) CreateSubject(c *fiber.Ctx) error {
	var subject domain.Subject

	err := c.BodyParser(&subject)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to create subject",
		})
	}

	err = uh.uc.CreateSubject(c.Context(), &subject)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to create subject",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subject successsfully added",
	})
}

func (uh *uHandler) CreateSubjectBulk(c *fiber.Ctx) error {
	var subjects []domain.Subject

	err := c.BodyParser(&subjects)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to create subject bulk",
		})
	}

	duplicateList, _ := uh.uc.CreateSubjectBulk(c.Context(), &subjects)
	if duplicateList != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   duplicateList,
			"success": false,
			"message": "Failed to create subject bulk",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successfully added",
	})
}

func (uh *uHandler) GetAllSubject(c *fiber.Ctx) error {
	datas, err := uh.uc.GetAllSubject(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to get all subject",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successsfully retrieved",
		"data":    datas,
	})
}

func (uh *uHandler) UpdateSubject(c *fiber.Ctx) error {
	var subject domain.Subject

	id := c.Params("id")
	subjectID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "String to Int Converter failure",
			"success": false,
			"message": "Failed to update subject",
		})
	}

	err = c.BodyParser(&subject)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to update subject",
		})
	}

	err = uh.uc.UpdateSubject(c.Context(), subjectID, &subject)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to update subject",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successsfully updated",
	})
}

func (uh *uHandler) DeleteSubject(c *fiber.Ctx) error {

	id := c.Params("id")
	subjectID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to delete subject",
		})
	}

	err = uh.uc.DeleteSubject(c.Context(), subjectID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to delete subject",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successsfully deleted",
	})
}

type CreateStaffRequest struct {
	User       domain.User `json:"user"`
	SubjectIDs []int       `json:"subject_ids"`
}

func (uh *uHandler) CreateStaff(c *fiber.Ctx) error {
	var req CreateStaffRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	req.User.Role = "staff"

	_, err := uh.uc.CreateStaff(c.Context(), &req.User, req.SubjectIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Account created successfully",
	})
}

func (uh *uHandler) GetAllStaff(c *fiber.Ctx) error {
	v, err := uh.uc.GetAllStaff(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff retrieved successfully",
		"data":    v,
	})
}

func (uh *uHandler) DeleteStaff(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
		})
	}

	err = uh.uc.DeleteStaff(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff deleted successfully",
	})
}

func (uh *uHandler) GetStaffDetail(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
		})
	}

	v, err := uh.uc.GetStaffDetail(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff retrieved successfully",
		"data":    v,
	})

}

func (uh *uHandler) ModifyStaff(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid staff ID",
		})
	}

	var payload struct {
		User       domain.User `json:"user"`
		SubjectIDs []int       `json:"subject_ids"` 
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	payload.User.Role = "staff"

	err = uh.uc.UpdateStaff(c.Context(), id, &payload.User, payload.SubjectIDs) 
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to modify staff",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff modified successfully",
	})
}
