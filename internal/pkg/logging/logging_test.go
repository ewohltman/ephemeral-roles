package logging

import (
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
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
	log := New("info", "America/New_York", "")
	log.SetOutput(ioutil.Discard)

	return log
}
