package config

import (
	"fmt"
	"os"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

func GetFiberListenAddress() string {
	return fmt.Sprintf("%s:%s", GetFiberHttpHost(), GetFiberHttpPort())
}

func GetFiberConfig() fiber.Config {
	return fiber.Config{
		DisableStartupMessage: false,
		JSONEncoder:           sonic.Marshal,
		JSONDecoder:           sonic.Unmarshal,
		Prefork:               false,
		ServerHeader:          "SINOAN",
		AppName:               GetAppName(),
		ReadTimeout:           time.Second * 60,
		CaseSensitive:         true,
	}
}

func GetAppName() string {
	v := os.Getenv("APP_NAME")
	if v == "" {
		return "SINOAN"
	}

	return v
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
