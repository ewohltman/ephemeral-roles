// Package logging provides a logrus logging implementation. Configuration
// is determined via environment variables upon startup and logging level may
// be changed at runtime.
package logging

import (
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
	WrappedLogger() *logrus.Logger
	UpdateLevel(level string)
}

// Logger is a struct to wrap a *logrus.Logger instance and provides custom
// methods.
type Logger struct {
	*logrus.Logger
	Location             *time.Location
	DiscordrusWebHookURL string
}

// New returns a new *Logger instance.
func New(variables *environment.Variables) *Logger {
	location := timestampLocation(variables.LogTimezoneLocation)

	log := &Logger{
		Logger: &logrus.Logger{
			Formatter: &localeFormatter{
				&logrus.TextFormatter{},
				location,
			},
			Out:   os.Stdout,
			Level: logrus.InfoLevel,
			Hooks: make(logrus.LevelHooks),
		},
		Location:             location,
		DiscordrusWebHookURL: variables.DiscordrusWebHookURL,
	}

	log.UpdateLevel(variables.LogLevel)

	return log
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

// UpdateLevel allows for runtime updates of the logging level.
func (log *Logger) UpdateLevel(level string) {
	// Update our global logging instance log level
	log.SetLevel(parseLevel(level))

	// Reset logging hooks
	log.Hooks = make(logrus.LevelHooks)

	// Check/apply `github.com/kz/discordrus` hook integration
	log.discordrusIntegration()
}

func (log *Logger) discordrusIntegration() {
	if log.DiscordrusWebHookURL == "" {
		return
	}

	timeString := time.Now().In(log.Location).String()
	timeZoneToken := strings.Split(timeString, " ")[3]

	log.AddHook(
		discordrus.NewHook(
			log.DiscordrusWebHookURL,
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
				TimestampLocale: log.Location,
			},
		),
	)
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

func timestampLocation(locationString string) *time.Location {
	location, err := time.LoadLocation(locationString)
	if err != nil {
		return time.UTC
	}

	return location
}

func parseLevel(level string) logrus.Level {
	logLevel := logrus.InfoLevel

	switch strings.ToLower(strings.TrimSpace(level)) {
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
