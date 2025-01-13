package middleware

import (
	"notification/domain"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte(os.Getenv("BYTE_KEY"))

func GenerateJWT(userID int, username, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &domain.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func RoleRequired(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userToken := c.Locals("user").(*domain.Claims)
		if userToken == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized: missing token",
			})
		}

		for _, role := range roles {
			if userToken.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden: insufficient permissions",
		})
	}
}

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenStr := c.Get("Authorization")
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing token",
			})
		}

		claims := new(domain.Claims)
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("user", claims)

		return c.Next()
	}
}
