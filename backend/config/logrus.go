package config

import "github.com/sirupsen/logrus"

var logrusInstance *logrus.Logger

func GetLogrusInstance() *logrus.Logger {
	if logrusInstance == nil {
		logrusInstance = logrus.New()
		logrusInstance.SetFormatter(&logrus.JSONFormatter{})
	}
	return logrusInstance
}
