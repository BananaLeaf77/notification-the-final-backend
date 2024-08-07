package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func GetFiberListenAddress() string {
	return fmt.Sprintf("%s:%s", GetFiberHttpHost(), GetFiberHttpPort())
}

func GetFiberConfig() fiber.Config {
	return fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	}
}

func GetFiberHttpHost() string {
	env := os.Getenv("HTTP_HOST")
	if env != "" {
		return env
	}
	return "0.0.0.0"
}

func GetFiberHttpPort() string {
	env := os.Getenv("HTTP_PORT")
	if env != "" {
		return env
	}
	return "8000"
}
