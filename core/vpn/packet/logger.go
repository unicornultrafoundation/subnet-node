package packet

import (
	"io"

	"github.com/sirupsen/logrus"
)

// SetupTestLogger sets up a test logger that discards all output
// Returns a function to restore the original logger
func SetupTestLogger() func() {
	// Save the original logger
	originalLogger := logrus.StandardLogger()
	originalLevel := originalLogger.GetLevel()
	originalOut := originalLogger.Out

	// Set up a new logger that discards all output
	logrus.SetOutput(io.Discard)
	logrus.SetLevel(logrus.ErrorLevel)

	// Return a function to restore the original logger
	return func() {
		logrus.SetOutput(originalOut)
		logrus.SetLevel(originalLevel)
	}
}
