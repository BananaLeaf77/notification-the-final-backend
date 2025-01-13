package delivery

import (
	"notification/config"
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

	group.Get("/show-student-testscores", handler.GetSubjectsForTeacher)
}

func NewUserHandlerDeploy(app *fiber.App, useCase domain.UserUseCase) {
	handler := &uHandler{
		uc: useCase,
	}
	group := app.Group("/user") // All routes under /user

	group.Post("/create-staff", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateStaff)
	group.Get("/get-all", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetAllStaff)
	group.Delete("/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStaff)
	group.Get("/details/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetStaffDetail)
	group.Put("/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.ModifyStaff)
	group.Post("/add-subject", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateSubject)
	group.Post("/add-subject-bulk", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.CreateSubjectBulk)
	group.Get("/subject/all", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.GetAllSubject)
	group.Put("/subject/modify/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.UpdateSubject)
	group.Delete("/subject/rm/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteSubject)
	group.Get("/show-user-assigned-subject", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.GetSubjectsForTeacher)
	group.Post("/input-test-scores", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.InputTestScores)
	group.Get("/profile-dashboard", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.ShowProfile)
	group.Post("/rm/users", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteStaffMass)
	group.Get("/subject/:id", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.GetSubjectDetail)
	group.Post("/rm/subjects", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.DeleteSubjectMass)
	group.Get("/get-all/test-scores", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.GetAllTestScores)
	group.Get("/get/test-scores/:subject_id", middleware.AuthRequired(), middleware.RoleRequired("admin", "staff"), handler.GetAllTestScoresBySubjectID)
	// group.Get("/reset/test-scores", middleware.AuthRequired(), middleware.RoleRequired("admin"), handler.ResetTestScore)
}

func (h *uHandler) GetAllTestScoresBySubjectID(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	subjectID := c.Params("subject_id")

	convertedSubjectID, err := strconv.Atoi(subjectID)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "GetAllTestScoresBySubjectID")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Converter failure on subject id",
			"success": false,
		})
	}

	data, err := h.uc.GetAllTestScoresBySubjectID(c.Context(), convertedSubjectID)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetAllTestScoresBySubjectID")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Failed to get all test score by subject id",
			"success": false,
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetAllTestScoresBySubjectID")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    data,
		"success": true,
		"message": "Test score retrieved successfully",
	})
}

func (h *uHandler) GetAllTestScores(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)

	datas, err := h.uc.GetAllTestScores(c.Context())
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetAllTestScores")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Get all Test Scores fail to deliver data",
			"data":    nil,
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetAllTestScores")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Test Score successsfully retrieved",
		"data":    datas,
	})
}

func (h *uHandler) DeleteSubjectMass(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	var payload struct {
		IDS []int `json:"subject_ids"`
	}

	err := c.BodyParser(&payload)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DeleteSubjectMass")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "parsing failure",
			"success": false,
		})
	}

	err = h.uc.DeleteSubjectMass(c.Context(), &payload.IDS)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DeleteSubjectMass")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to mass delete subject",
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DeleteSubjectMass")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects deleted successfully",
	})

}

func (h *uHandler) GetSubjectDetail(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	subjectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "GetSubjectDetail")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
		})
	}

	v, err := h.uc.GetSubjectDetail(c.Context(), subjectID)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "GetSubjectDetail")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get subject detail",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "GetSubjectDetail")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subject detail retrieved successfully",
		"data":    v,
	})
}

func (h *uHandler) DeleteStaffMass(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	var payload struct {
		IDS []int `json:"ids"`
	}

	err := c.BodyParser(&payload)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusBadRequest, "DeleteStaffMass")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete users",
			"error":   err.Error(),
		})
	}

	err = h.uc.DeleteStaffMass(c.Context(), &payload.IDS)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "DeleteStaffMass")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete users",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "DeleteStaffMass")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Users deleted successfully",
	})
}

func (h *uHandler) ShowProfile(c *fiber.Ctx) error {
	userToken := c.Locals("user").(*domain.Claims)
	v, err := h.uc.ShowProfile(c.Context(), userToken.UserID)
	if err != nil {
		config.PrintLogInfo(&userToken.Username, fiber.StatusInternalServerError, "ShowProfile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Failed to retrieved profile data",
			"data":    nil,
		})
	}

	config.PrintLogInfo(&userToken.Username, fiber.StatusOK, "ShowProfile")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile data loaded",
		"data":    v,
	})
}

func (h *uHandler) InputTestScores(c *fiber.Ctx) error {
	userClaims, _ := c.Locals("user").(*domain.Claims)

	teacherID := userClaims.UserID
	// var testScores []domain.TestScore
	var thePayload domain.InputTestScorePayload
	if err := c.BodyParser(&thePayload); err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "InputTestScores")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request payload"})
	}

	err := h.uc.InputTestScores(c.Context(), teacherID, &thePayload)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "InputTestScores")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to input test scores",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "InputTestScores")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Test scores successfully inputted",
	})
}

func (h *uHandler) GetSubjectsForTeacher(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	userID := userClaims.UserID

	subjects, err := h.uc.GetSubjectsForTeacher(c.Context(), userID)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "GetSubjectsForTeacher")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to get subjects for the teacher",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Subjects fetched successfully",
		"data":    subjects,
	})
}

func (uh *uHandler) CreateSubject(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)
	var subject domain.Subject

	err := c.BodyParser(&subject)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "CreateSubject")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to create subject",
		})
	}

	err = uh.uc.CreateSubject(c.Context(), &subject)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "CreateSubject")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to create subject",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "CreateSubject")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subject successsfully added",
	})
}

func (uh *uHandler) CreateSubjectBulk(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	var subjects []domain.Subject

	err := c.BodyParser(&subjects)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "CreateSubjectBulk")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to create subject bulk",
		})
	}

	duplicateList, _ := uh.uc.CreateSubjectBulk(c.Context(), &subjects)
	if duplicateList != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "CreateSubjectBulk")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   duplicateList,
			"success": false,
			"message": "Failed to create subject bulk",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "CreateSubjectBulk")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successfully added",
	})
}

func (uh *uHandler) GetAllSubject(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	datas, err := uh.uc.GetAllSubject(c.Context(), userClaims.UserID)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "GetAllSubject")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to get all subject",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "GetAllSubject")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successsfully retrieved",
		"data":    datas,
	})
}

func (uh *uHandler) UpdateSubject(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)
	var subject domain.Subject

	id := c.Params("id")
	subjectID, err := strconv.Atoi(id)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "UpdateSubject")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "String to Int Converter failure on Subject ID",
			"success": false,
			"message": "Failed to update subject",
		})
	}

	err = c.BodyParser(&subject)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "UpdateSubject")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to update subject",
		})
	}

	err = uh.uc.UpdateSubject(c.Context(), subjectID, &subject)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "UpdateSubject")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to update subject",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "UpdateSubject")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Subjects successsfully updated",
	})
}

func (uh *uHandler) DeleteSubject(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	id := c.Params("id")
	subjectID, err := strconv.Atoi(id)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "DeleteSubject")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to delete subject",
		})
	}

	err = uh.uc.DeleteSubject(c.Context(), subjectID)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "DeleteSubject")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to delete subject",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "DeleteSubject")
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
	userClaims := c.Locals("user").(*domain.Claims)
	var req domain.User

	if err := c.BodyParser(&req); err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "CreateStaff")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	if req.Username == "" || req.Password == "" || req.Name == "" {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "CreateStaff")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid data",
			"success": false,
		})
	}

	_, err := uh.uc.CreateStaff(c.Context(), &req)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "CreateStaff")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "CreateStaff")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Account created successfully",
	})
}

func (uh *uHandler) GetAllStaff(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	v, err := uh.uc.GetAllStaff(c.Context())
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "GetAllStaff")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "GetAllStaff")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff retrieved successfully",
		"data":    v,
	})
}

func (uh *uHandler) DeleteStaff(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "DeleteStaff")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
		})
	}

	err = uh.uc.DeleteStaff(c.Context(), id)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "DeleteStaff")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "DeleteStaff")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff deleted successfully",
	})
}

func (uh *uHandler) GetStaffDetail(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)
	id, err := strconv.Atoi(c.Params("id"))
	if userClaims.UserID == 1 && id == 1 && userClaims.Role == "admin" {
		v, err := uh.uc.GetAdminByAdmin(c.Context())
		if err != nil {
			config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "GetAdminByAdmin")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   err.Error(),
				"success": false,
				"message": "Failed to retrieved staff data",
			})
		}

		config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "GetAdminByAdmin")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    v,
		})
	}

	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "GetStaffDetail")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "converter failure",
			"success": false,
			"message": "Failed to retrieved staff data",
		})
	}

	v, err := uh.uc.GetStaffDetail(c.Context(), id)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "GetStaffDetail")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   err.Error(),
			"success": false,
			"message": "Failed to retrieved staff data",
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "GetStaffDetail")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff retrieved successfully",
		"data":    v,
	})

}

func (uh *uHandler) ModifyStaff(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(*domain.Claims)

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "ModifyStaff")
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
		config.PrintLogInfo(&userClaims.Username, fiber.StatusBadRequest, "ModifyStaff")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	err = uh.uc.UpdateStaff(c.Context(), id, &payload.User, payload.SubjectIDs)
	if err != nil {
		config.PrintLogInfo(&userClaims.Username, fiber.StatusInternalServerError, "ModifyStaff")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to modify staff",
			"error":   err.Error(),
		})
	}

	config.PrintLogInfo(&userClaims.Username, fiber.StatusOK, "ModifyStaff")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Staff modified successfully",
	})
}
