package logging_test

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kz/discordrus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

const updateError = "Failed update logging level"

func TestNew(t *testing.T) {
	t.Parallel()

	testLogger()
}

func TestLogger_WrappedLogger(t *testing.T) {
	t.Parallel()

	log := testLogger().WrappedLogger()

	require.NotNil(t, log)
}

func TestLogger_UpdateLevel(t *testing.T) {
	t.Parallel()

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

		assert.Equal(t, testLevel, log.Logger.Level, updateError)
	}

	log.DiscordrusWebHookURL = ""

	log.UpdateLevel(logrus.DebugLevel.String())

	assert.Equal(t, logrus.DebugLevel, log.Logger.Level, updateError)
}

func TestLogger_UpdateDiscordrus(t *testing.T) {
	t.Parallel()

	const expected = "updateTest"

	log := testLogger()

	log.DiscordrusWebHookURL = ""
	log.UpdateDiscordrus()

	assert.Empty(t, log.Logger.Hooks)

	log.DiscordrusWebHookURL = expected
	log.UpdateDiscordrus()

	hook := log.Logger.Hooks[logrus.InfoLevel][0].(*discordrus.Hook)

	assert.Equal(t, expected, hook.WebhookURL)
}

func TestLogger_DiscordGoLogf(t *testing.T) {
	t.Parallel()

	log := testLogger()
	log.DiscordrusWebHookURL = ""
	log.UpdateDiscordrus()
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
	t.Parallel()

	const expectedFormat = `time="0001-01-01T00:00:00Z" level=panic shardID=0`

	log := testLogger()

	locale := &logging.Locale{
		Location:  nil,
		Formatter: &logrus.TextFormatter{},
	}

	_, err := locale.Format(log.Entry)
	require.NoError(t, err)

	locale.Location = time.UTC

	actualFormat, err := locale.Format(log.Entry)
	require.NoError(t, err)

	actualFormatString := strings.TrimSpace(string(actualFormat))

	assert.Equal(t, expectedFormat, actualFormatString)
}

func testLogger() *logging.Logger {
	return logging.New(
		logging.OptionalOutput(io.Discard),
		logging.OptionalShardID(0),
		logging.OptionalLogLevel("info"),
		logging.OptionalTimezoneLocation("xyz"),
		logging.OptionalTimezoneLocation("America/New_York"),
		logging.OptionalDiscordrus("test"),
	)
}
