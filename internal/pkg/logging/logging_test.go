package logging

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/environment"
)

const updateError = "Failed update logging level"

func TestNew(t *testing.T) {
	testLogger()
}

func TestLogger_UpdateLevel(t *testing.T) {
	log := testLogger()

	testLevels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	for _, testLevel := range testLevels {
		log.UpdateLevel(testLevel.String())

		if log.Level != testLevel {
			t.Error(updateError)
		}
	}
}

func TestLogger_WrappedLogger(t *testing.T) {
	log := testLogger().WrappedLogger()

	if log == nil {
		t.Fatal("Unexpected nil wrapped *logrus.Logger")
	}
}

func testLogger() *Logger {
	variables := &environment.Variables{
		LogLevel:             "info",
		LogTimezoneLocation:  "America/New_York",
		DiscordrusWebHookURL: "",
	}

	log := New(variables)
	log.SetOutput(ioutil.Discard)

	return log
}
