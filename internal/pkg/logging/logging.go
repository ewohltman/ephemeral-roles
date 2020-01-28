// Package logging provides a logrus logging implementation. Configuration
// is determined via environment variables upon startup and logging level may
// be changed at runtime.
package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kz/discordrus"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/environment"
)

// Logging level strings.
const (
	DebugLevel   = "debug"
	InfoLevel    = "info"
	WarningLevel = "warning"
	ErrorLevel   = "error"
	FatalLevel   = "fatal"
	PanicLevel   = "panic"
)

// Logging level color constants.
const (
	DebugColor   = 10170623
	InfoColor    = 3581519
	WarningColor = 14327864
	ErrorColor   = 13631488
	PanicColor   = 13631488
	FatalColor   = 13631488
)

// Interface wraps the logrus.FieldLogger interface and includes custom
// methods.
type Interface interface {
	logrus.FieldLogger
	UpdateLevel()
	WrappedLogger() *logrus.Logger
}

// Logger is a struct to wrap a *logrus.Logger instance and provides custom
// methods.
type Logger struct {
	*logrus.Logger
}

// New returns a new *logrus.Logger instance.
func New() *Logger {
	timestampLocale, err := timeLocalization()
	if err != nil {
		timestampLocale = time.Local
	}

	log := &Logger{
		Logger: &logrus.Logger{
			Formatter: &localeFormatter{
				&logrus.TextFormatter{},
				timestampLocale,
			},
			Out:   os.Stdout,
			Level: logrus.InfoLevel,
			Hooks: make(logrus.LevelHooks),
		},
	}

	log.UpdateLevel()

	return log
}

// UpdateLevel allows for runtime updates of the logging level and resets the
// hooks with new values from the environment.
func (log *Logger) UpdateLevel() {
	// Update our global logging instance log level
	log.SetLevel(environmentLevel())

	// Reset logging hooks
	log.Hooks = make(logrus.LevelHooks)

	// Check/apply `github.com/kz/discordrus` hook integration
	discordrusIntegration(log)
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

type localeFormatter struct {
	logrus.Formatter
	*time.Location
}

// Format satisfies the logrus.Formatter interface.
func (l *localeFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.In(l.Location)

	return l.Formatter.Format(e)
}

func timeLocalization() (*time.Location, error) {
	envLocation, found := os.LookupEnv(environment.LogTimezoneLocation)
	if !found || envLocation == "" {
		return time.Local, nil
	}

	timeLocalization, err := time.LoadLocation(envLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to load location %s: %s", environment.LogTimezoneLocation, err)
	}

	return timeLocalization, nil
}

func environmentLevel() logrus.Level {
	logLevel := logrus.InfoLevel // Default to InfoLevel

	envLevel, found := os.LookupEnv(environment.LogLevel)
	if !found || envLevel == "" {
		return logLevel
	}

	switch strings.ToLower(strings.TrimSpace(envLevel)) {
	case DebugLevel:
		logLevel = logrus.DebugLevel
	case InfoLevel:
		logLevel = logrus.InfoLevel
	case WarningLevel:
		logLevel = logrus.WarnLevel
	case ErrorLevel:
		logLevel = logrus.ErrorLevel
	case FatalLevel:
		logLevel = logrus.FatalLevel
	case PanicLevel:
		logLevel = logrus.PanicLevel
	}

	return logLevel
}

func discordrusIntegration(log *Logger) {
	if hookURLString, found := os.LookupEnv(environment.DiscordrusWebHookURL); found {
		timeString := ""

		timestampLocale, err := timeLocalization()
		if err != nil {
			log.WithError(err).Debugf("Unable to determine timestamp locale, defaulting to local runtime")

			timeString = time.Now().String()
		} else {
			timeString = time.Now().In(timestampLocale).String()
		}

		timeZoneToken := strings.Split(timeString, " ")[3]

		log.AddHook(
			discordrus.NewHook(
				hookURLString,
				log.Level,
				&discordrus.Opts{
					Username:           "",
					Author:             "",
					EnableCustomColors: true,
					CustomLevelColors: &discordrus.LevelColors{
						Debug: DebugColor,
						Info:  InfoColor,
						Warn:  WarningColor,
						Error: ErrorColor,
						Panic: PanicColor,
						Fatal: FatalColor,
					},
					TimestampFormat: "Jan 2 15:04:05.00000 " + timeZoneToken,
					TimestampLocale: timestampLocale,
				},
			),
		)
	}
}
