package config

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v2/log"
	"github.com/sirupsen/logrus"
)

var logrusInstance *logrus.Logger

func GetLogrusInstance() *logrus.Logger {
	if logrusInstance == nil {
		logrusInstance = logrus.New()
		logrusInstance.SetFormatter(&logrus.JSONFormatter{})
	}
	return logrusInstance
}

const (
	green  = "\033[32m" // Green for 200 OK
	yellow = "\033[33m" // Yellow for 300 series
	red    = "\033[31m" // Red for 400 and 500 series
	reset  = "\033[0m"  // Reset to default color
)

func PrintLogInfo(username *string, statusCode int, functionName string) {
	var logColor string

	switch statusCode {
	case fiber.StatusOK:
		logColor = green
	case fiber.StatusCreated:
		logColor = green
	case fiber.StatusAccepted:
		logColor = yellow
	case fiber.StatusBadRequest:
		logColor = red
	case fiber.StatusUnauthorized:
		logColor = red
	case fiber.StatusInternalServerError:
		logColor = red
	default:
		logColor = reset // Default color for any unhandled status codes
	}

	// Log the message with the appropriate color
	logMsg := fmt.Sprintf("\n\n\nUser: %s, (%s) => Status: %s[%d] - %s%s\n\n\n", *username, functionName, logColor, statusCode, http.StatusText(statusCode), reset)
	log.Info(logMsg)
}
