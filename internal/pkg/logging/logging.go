// Package logging should be used by other packages to get a pointer to the
// global logrus logging instance via the Instance function call
//
// Logging configuration is determined via environment variables upon startup.
// These variables may be manipulated during runtime and then reflected in the
// global logging instance via the UpdateLevel function call
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
	WrappedLogger() *logrus.Logger
	UpdateLevel()
}

// Logger is a struct to wrap a *logrus.Logger instance and provide custom
// methods..
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

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
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

type localeFormatter struct {
	logrus.Formatter
	*time.Location
}

// Format satisfies the logrus.Formatter interface.
func (l *localeFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.In(l.Location)

	return l.Formatter.Format(e)
}

// timeLocalization returns a *time.Location defined in environment variables,
// or otherwise defaults to time.Local.
func timeLocalization() (*time.Location, error) {
	envLocation, found := os.LookupEnv(environment.LogTimezoneLocation)
	if !found || envLocation == "" {
		envLocation = time.Local.String()
	}

	timeLocalization, err := time.LoadLocation(envLocation)
	if err != nil {
		return nil, fmt.Errorf("unable to load location %s: %s", environment.LogTimezoneLocation, err)
	}

	return timeLocalization, nil
}

// environmentLevel parses and returns our logging level from the environment.
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

// discordrusIntegration checks to see if we can apply an optional integration
// support for a `github.com/kz/discordrus` hook.
func discordrusIntegration(log *Logger) {
	if hookURLString, found := os.LookupEnv(environment.DiscordrusWebHookURL); found {
		timeString := ""

		timestampLocale, err := timeLocalization()
		if err != nil {
			log.WithError(err).Debugf("Unable to determine timestamp locale, defaulting to local runtime")

			timeString = time.Now().String()
		} else {
			log.WithField("locale", timestampLocale.String()).Debugf("Found custom logging timestamp locale")

			timeString = time.Now().In(timestampLocale).String()
		}

		// timeZoneTokens => [2017-12-23] [11:45:53.0703314] [-0000] [UTC]
		timeZoneToken := strings.Split(timeString, " ")[3]

		timeStampFormat := "Jan 2 15:04:05.00000 " + timeZoneToken

		log.AddHook(
			discordrus.NewHook(
				hookURLString,
				log.Level,
				&discordrus.Opts{
					Username:            "",
					Author:              "",    // Setting this to a non-empty string adds the author text to the message header
					DisableInlineFields: false, // If set to true, fields will not appear in columns ("inline")
					EnableCustomColors:  true,  // If set to true, the below CustomLevelColors will apply
					CustomLevelColors: &discordrus.LevelColors{
						Debug: DebugColor,
						Info:  InfoColor,
						Warn:  WarningColor,
						Error: ErrorColor,
						Panic: PanicColor,
						Fatal: FatalColor,
					},
					DisableTimestamp: false,           // Setting this to true will disable timestamps from appearing in the footer
					TimestampFormat:  timeStampFormat, // The timestamp takes this format; if unset, it will take a default format
					TimestampLocale:  timestampLocale, // The timestamp takes it's timezone from the provided locale
				},
			),
		)
	}
}
