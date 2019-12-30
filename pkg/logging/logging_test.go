package logging

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func testLogger() *logrus.Logger {
	log := New()
	log.SetOutput(ioutil.Discard)

	return log
}

func TestNew(t *testing.T) {
	testLogger()
}

func TestReinitialize(t *testing.T) {
	testLog := testLogger()

	originalLevel := os.Getenv("LOG_LEVEL")

	err := os.Setenv("LOG_LEVEL", "debug")
	if err != nil {
		t.Error(err)
	}

	UpdateLevel(testLog)

	if testLog.Level != logrus.DebugLevel {
		t.Errorf("Failed runtime logging reinitialization")
	}

	err = os.Setenv("LOG_LEVEL", "info")
	if err != nil {
		t.Error(err)
	}

	UpdateLevel(testLog)

	if testLog.Level != logrus.InfoLevel {
		t.Errorf("Failed runtime logging reinitialization")
	}

	err = os.Setenv("LOG_LEVEL", originalLevel)
	if err != nil {
		t.Error(err)
	}

	UpdateLevel(testLog)
}
