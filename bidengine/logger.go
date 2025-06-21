package bidengine

import (
	"os"

	"github.com/sirupsen/logrus"
)

// LoggerService implements Logger interface using logrus
type LoggerService struct {
	logger *logrus.Logger
}

// NewLogger creates a new LoggerService instance using logrus
func NewLogger(level, logFile string) (*LoggerService, error) {
	logger := logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Set output
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		logger.SetOutput(file)
	} else {
		logger.SetOutput(os.Stdout)
	}

	// Set formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	return &LoggerService{
		logger: logger,
	}, nil
}

// Debug logs a debug message
func (l *LoggerService) Debug(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Debug(msg)
}

// Info logs an info message
func (l *LoggerService) Info(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Info(msg)
}

// Warn logs a warning message
func (l *LoggerService) Warn(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Warn(msg)
}

// Error logs an error message
func (l *LoggerService) Error(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Error(msg)
}

// Fatal logs a fatal message and exits
func (l *LoggerService) Fatal(msg string, fields ...interface{}) {
	l.logger.WithFields(l.parseFields(fields...)).Fatal(msg)
}

// parseFields converts variadic interface{} arguments to logrus.Fields
func (l *LoggerService) parseFields(fields ...interface{}) logrus.Fields {
	result := make(logrus.Fields)

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if ok {
				result[key] = fields[i+1]
			}
		}
	}

	return result
}
