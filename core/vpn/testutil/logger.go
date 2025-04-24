package testutil

import (
	"io"
	"log"
)

// SetupTestLogger sets up a test logger that discards all output
func SetupTestLogger() func() {
	// Save the original logger settings
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()

	// Create a new logger that discards all output
	log.SetOutput(io.Discard)

	// Return a function to restore the original logger
	return func() {
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	}
}
