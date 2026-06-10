// Package logger configures application logging for Adele built on logrus.
//
// It creates loggers whose format and level are driven by environment variables and
// provides structured HTTP request logging middleware for incoming requests.
package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Configure and initialize a new Logrus logger instance based on environment variables.
func CreateLogger() *logrus.Logger {

	log := logrus.New()

	if os.Getenv("LOG_FORMAT") == "JSON" {
		log.SetFormatter(&logrus.JSONFormatter{})
	} else {
		log.SetFormatter(&logrus.TextFormatter{})
	}

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(logrus.DebugLevel)
	} else if os.Getenv("LOG_LEVEL") != "" {
		log.SetLevel(GetLogLevel(os.Getenv("LOG_LEVEL")))
	}

	return log
}

// Convert a string value into the corresponding logrus.Level type where the default
// value is a info log level - supports various logging levels, like Trace, Debug, Info,
// Warning, Error, Fatal, and Panic, in increasing order of severity.
func GetLogLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warn", "warning":
		return logrus.WarnLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}
