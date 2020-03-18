package logging

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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

		if log.Logger.Level != testLevel {
			t.Error(updateError)
		}
	}
}

func TestLocale_Format(t *testing.T) {
	const expectedFormat = `time="0001-01-01T00:00:00Z" level=panic`

	log := testLogger()

	entry := logrus.NewEntry(log.Logger)

	locale := &locale{
		&logrus.TextFormatter{},
		time.UTC,
	}

	actualFormat, err := locale.Format(entry)
	if err != nil {
		t.Fatalf("Error formating entry: %s", err)
	}

	actualFormatString := strings.TrimSpace(string(actualFormat))

	if actualFormatString != expectedFormat {
		t.Fatalf(
			"Unexpected format. Got: %s, Expected: %s",
			string(actualFormat),
			expectedFormat,
		)
	}
}

func testLogger() *Logger {
	log := New(0, "info", "America/New_York", "test")
	log.Logger.SetOutput(ioutil.Discard)

	return log
}
