package packet

import (
	"io"

	"github.com/sirupsen/logrus"
)

// TestLogLevel controls the log level during tests
var TestLogLevel = logrus.WarnLevel

// SetupTestLogger configures a test-specific logger that reduces verbosity
func SetupTestLogger() func() {
	// Get the global logger instance
	logger := logrus.StandardLogger()

	// Save the original log level and output
	originalLevel := logger.GetLevel()
	originalOutput := logger.Out

	// Set a higher log level to reduce verbosity during tests
	logger.SetLevel(TestLogLevel)

	// Optionally discard logs completely
	if TestLogLevel == logrus.PanicLevel {
		logger.SetOutput(io.Discard)
	}

	// Return a function to restore the original logger settings
	return func() {
		logger.SetLevel(originalLevel)
		logger.SetOutput(originalOutput)
	}
}
