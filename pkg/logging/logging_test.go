package logging

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInstance(t *testing.T) {
	testLog := Instance()

	if testLog == nil {
		t.Errorf("Failed obtaining global logging instance")
	}
}

func TestReinitialize(t *testing.T) {
	originalLevel := os.Getenv("LOG_LEVEL")
	t.Log("LOG_LEVEL (original): " + originalLevel)

	testLog := Instance()

	err := os.Setenv("LOG_LEVEL", "debug")
	if err != nil {
		t.Error(err)
	}
	t.Log("LOG_LEVEL: " + os.Getenv("LOG_LEVEL"))
	Reinitialize()

	if testLog.Level != logrus.DebugLevel {
		t.Errorf("Failed runtime logging reinitialization")
	}

	err = os.Setenv("LOG_LEVEL", "info")
	if err != nil {
		t.Error(err)
	}
	t.Log("LOG_LEVEL: " + os.Getenv("LOG_LEVEL"))
	Reinitialize()

	if testLog.Level != logrus.InfoLevel {
		t.Errorf("Failed runtime logging reinitialization")
	}

	err = os.Setenv("LOG_LEVEL", originalLevel)
	if err != nil {
		t.Error(err)
	}
	t.Log("LOG_LEVEL: " + os.Getenv("LOG_LEVEL"))
	Reinitialize()
}
