package config

import (
	"encoding/json"
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
	case fiber.StatusOK, fiber.StatusCreated:
		logColor = green
	case fiber.StatusAccepted:
		logColor = yellow
	case fiber.StatusBadRequest, fiber.StatusUnauthorized, fiber.StatusInternalServerError:
		logColor = red
	default:
		logColor = reset
	}

	// Handle a nil `username` by using a placeholder
	user := "Unknown"
	if username != nil {
		user = *username
	}

	logMsg := fmt.Sprintf("\nUser: %s, (%s) => Status: %s[%d] - %s%s\n\n\n", user, functionName, logColor, statusCode, http.StatusText(statusCode), reset)
	log.Info(logMsg)
}

func PrintStruct(strck interface{}) {
	// Marshal the struct into JSON
	jsonData, err := json.Marshal(strck)
	if err != nil {
		fmt.Println("Error marshaling struct:", err)
		return
	}

	// Print the JSON string representation of the struct
	fmt.Println(string(jsonData))
}
