package logging

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kz/discordrus"
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

	log.DiscordrusWebHookURL = ""

	log.UpdateLevel(logrus.DebugLevel.String())

	if log.Logger.Level != logrus.DebugLevel {
		t.Error(updateError)
	}
}

func TestLogger_UpdateDiscordrus(t *testing.T) {
	const expected = "updateTest"

	log := testLogger()

	log.DiscordrusWebHookURL = ""
	log.UpdateDiscordrus()

	if len(log.Logger.Hooks) != 0 {
		t.Errorf("Unexpected number of hooks: %d", len(log.Logger.Hooks))
	}

	log.DiscordrusWebHookURL = expected
	log.UpdateDiscordrus()

	hook := log.Logger.Hooks[logrus.InfoLevel][0].(*discordrus.Hook)

	if hook.WebhookURL != expected {
		t.Errorf(
			"Unexpected webhook URL: %s, expected: %s",
			hook.WebhookURL,
			expected,
		)
	}
}

func TestLogger_DiscordGoLogf(t *testing.T) {
	log := testLogger()
	log.DiscordrusWebHookURL = ""
	log.UpdateLevel(logrus.InfoLevel.String())

	logLevels := []int{
		discordgo.LogError,
		discordgo.LogWarning,
		discordgo.LogInformational,
		discordgo.LogDebug,
	}

	for _, logLevel := range logLevels {
		log.DiscordGoLogf(logLevel, 0, "Test: %d", 123)
	}
}

func TestLocale_Format(t *testing.T) {
	const expectedFormat = `time="0001-01-01T00:00:00Z" level=panic shardID=0`

	log := testLogger()

	locale := &locale{
		Location:  nil,
		Formatter: &logrus.TextFormatter{},
	}

	_, err := locale.Format(log.Entry)
	if err != nil {
		t.Errorf("Error formating entry: %s", err)
	}

	locale.Location = time.UTC

	actualFormat, err := locale.Format(log.Entry)
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
	log := New(
		OptionalShardID(0),
		OptionalLogLevel("info"),
		OptionalTimezoneLocation("xyz"),
		OptionalTimezoneLocation("America/New_York"),
		OptionalDiscordrus("test"),
	)

	log.Logger.SetOutput(ioutil.Discard)

	return log
}
