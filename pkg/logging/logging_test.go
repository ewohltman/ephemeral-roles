package logging

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/environment"
)

const updateError = "Failed update logging level"

func TestNew(t *testing.T) {
	testLogger()
}

func TestLogger_WrappedLogger(t *testing.T) {
	log := testLogger().WrappedLogger()

	if log == nil {
		t.Fatal("Unexpected nil wrapped *logrus.Logger")
	}
}

func TestLogger_UpdateLevel(t *testing.T) {
	log := testLogger()

	originalLevel := os.Getenv(environment.LogLevel)

	testLevels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	for _, testLevel := range testLevels {
		err := changeLogLevel(log, testLevel.String())
		if err != nil {
			t.Error(err)
		}

		if log.Level != testLevel {
			t.Error(updateError)
		}
	}

	err := os.Setenv(environment.LogLevel, originalLevel)
	if err != nil {
		t.Fatalf("Unable to reset environment variable %s", environment.LogLevel)
	}
}

func testLogger() *Logger {
	log := New()
	log.SetOutput(ioutil.Discard)

	return log
}

func changeLogLevel(log *Logger, logLevel string) error {
	err := os.Setenv(environment.LogLevel, logLevel)
	if err != nil {
		return err
	}

	log.UpdateLevel()

	return nil
}
