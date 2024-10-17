package delivery

import (
	"fmt"
	"notification/domain"
	"notification/middleware"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type userHandler struct {
	db *gorm.DB
}

func NewUserAuthHandler(app *fiber.App, db *gorm.DB) {
	handler := &userHandler{
		db: db,
	}

	route := app.Group("/login")
	route.Post("/user", handler.Login)
}

func (h *userHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	var user domain.User
	err := h.db.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid username or password",
		})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		fmt.Println("Password comparison failed")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid username or password",
		})
	}

	token, err := middleware.GenerateJWT(user.UserID, user.Username, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(domain.LoginResponse{
		Token: token,
		Role:  user.Role,
	})
}
