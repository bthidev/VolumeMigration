package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Set default configuration
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
}

// GetLogger returns the singleton logger instance used throughout the application.
// The logger is configured with timestamp formatting and defaults to Info level.
// Use SetVerbose to change the logging level at runtime.
func GetLogger() *logrus.Logger {
	return log
}

// SetVerbose controls the logging verbosity level.
// When verbose is true, the logger is set to Debug level, displaying detailed
// diagnostic information useful for troubleshooting. When false, the logger
// uses Info level, showing only important operational messages.
// This should be called early in application startup based on user flags.
func SetVerbose(verbose bool) {
	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}
