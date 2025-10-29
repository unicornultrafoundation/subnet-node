package testutil

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func Logger(t testing.TB) *logrus.Logger {
	return logrus.New().WithField("test", t.Name()).Logger
}
