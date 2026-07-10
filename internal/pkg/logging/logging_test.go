package logging_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

func TestNew(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	log := logging.New(logging.OptionalOutput(buf))

	require.NotNil(t, log)

	log.Info("hello world")

	assert.Contains(t, buf.String(), "hello world")
}

func TestLogger_UpdateLevel(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	log := logging.New(
		logging.OptionalOutput(buf),
		logging.OptionalLogLevel(logging.InfoLevel),
	)

	log.Debug("debug-hidden")
	assert.NotContains(t, buf.String(), "debug-hidden")

	log.UpdateLevel(logging.DebugLevel)

	log.Debug("debug-shown")
	assert.Contains(t, buf.String(), "debug-shown")
}

func TestLogger_LevelParsing(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		logLevel string
		belowMsg string
		belowFn  func(log *logging.Logger, msg string)
		atMsg    string
		atFn     func(log *logging.Logger, msg string)
	}{
		{
			name:     "debug emits debug",
			logLevel: logging.DebugLevel,
			atMsg:    "debug-at",
			atFn:     func(log *logging.Logger, msg string) { log.Debug(msg) },
		},
		{
			name:     "warning suppresses info",
			logLevel: logging.WarningLevel,
			belowMsg: "info-below",
			belowFn:  func(log *logging.Logger, msg string) { log.Info(msg) },
			atMsg:    "warn-at",
			atFn:     func(log *logging.Logger, msg string) { log.Warn(msg) },
		},
		{
			name:     "error suppresses warn",
			logLevel: logging.ErrorLevel,
			belowMsg: "warn-below",
			belowFn:  func(log *logging.Logger, msg string) { log.Warn(msg) },
			atMsg:    "error-at",
			atFn:     func(log *logging.Logger, msg string) { log.Error(msg) },
		},
		{
			name:     "fatal maps to error and suppresses warn",
			logLevel: logging.FatalLevel,
			belowMsg: "warn-below-fatal",
			belowFn:  func(log *logging.Logger, msg string) { log.Warn(msg) },
			atMsg:    "error-at-fatal",
			atFn:     func(log *logging.Logger, msg string) { log.Error(msg) },
		},
		{
			name:     "unknown defaults to info",
			logLevel: "bogus",
			belowMsg: "debug-below",
			belowFn:  func(log *logging.Logger, msg string) { log.Debug(msg) },
			atMsg:    "info-at",
			atFn:     func(log *logging.Logger, msg string) { log.Info(msg) },
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			log := logging.New(
				logging.OptionalOutput(buf),
				logging.OptionalLogLevel(testCase.logLevel),
			)

			if testCase.belowFn != nil {
				testCase.belowFn(log, testCase.belowMsg)
				assert.NotContains(t, buf.String(), testCase.belowMsg)
			}

			testCase.atFn(log, testCase.atMsg)
			assert.Contains(t, buf.String(), testCase.atMsg)
		})
	}
}

func TestLogger_ShardIDAndTimezone(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	log := logging.New(
		logging.OptionalOutput(buf),
		logging.OptionalShardID(7),
		logging.OptionalTimezoneLocation("America/New_York"),
	)

	log.Info("hello")

	out := buf.String()
	assert.Contains(t, out, "shardID=7")
	assert.NotContains(t, out, "error parsing timezone location")
}

func TestLogger_InvalidTimezoneWarns(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}

	logging.New(
		logging.OptionalOutput(buf),
		logging.OptionalTimezoneLocation("not-a-timezone"),
	)

	out := buf.String()
	assert.Contains(t, out, "error parsing timezone location")
	assert.Contains(t, out, "not-a-timezone")
}

func TestLogger_FanoutToStdoutAndDiscord(t *testing.T) {
	t.Parallel()

	received := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		select {
		case received <- string(body):
		default:
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	buf := &bytes.Buffer{}
	log := logging.New(
		logging.OptionalOutput(buf),
		logging.OptionalLogLevel(logging.InfoLevel),
		logging.OptionalDiscordWebhook(server.URL),
	)

	log.Info("fanout-message", "marker", "unique-marker-value")

	assert.Contains(t, buf.String(), "fanout-message")
	assert.Contains(t, buf.String(), "unique-marker-value")

	select {
	case body := <-received:
		assert.Contains(t, body, "fanout-message")
		assert.Contains(t, body, "unique-marker-value")
	case <-time.After(2 * time.Second):
		require.Fail(t, "discord webhook did not receive the log record")
	}
}

func TestLogger_DiscordRespectsInfoLevel(t *testing.T) {
	t.Parallel()

	received := make(chan string, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		select {
		case received <- string(body):
		default:
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	buf := &bytes.Buffer{}
	log := logging.New(
		logging.OptionalOutput(buf),
		logging.OptionalLogLevel(logging.InfoLevel),
		logging.OptionalDiscordWebhook(server.URL),
	)

	log.Debug("debug-should-not-ship")

	assert.NotContains(t, buf.String(), "debug-should-not-ship")

	select {
	case body := <-received:
		require.Failf(t, "debug record leaked to Discord at info level", "body: %s", body)
	case <-time.After(500 * time.Millisecond):
	}
}
