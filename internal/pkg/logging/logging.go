// Package logging provides a logrus logging implementation. Configuration
// is determined via environment variables upon startup and logging level may
// be changed at runtime.
package logging

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kz/discordrus"
	"github.com/sirupsen/logrus"
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
	UpdateDiscordrus()
	DiscordGoLogf(discordgoLevel, caller int, format string, arguments ...interface{})
}

// OptionFunc is used to configure options for a *Logger.
type OptionFunc func(*Logger)

// Logger wraps a *logrus.Logger instance and provides custom methods.
type Logger struct {
	sync.Mutex
	*logrus.Entry
	Location             *time.Location
	DiscordrusWebHookURL string
}

// New returns a new *Logger instance configured with the OptionFunc arguments
// provided.
func New(options ...OptionFunc) *Logger {
	localeFormatter := &locale{
		Location:  time.UTC,
		Formatter: &logrus.TextFormatter{},
	}

	logger := &Logger{
		Entry: logrus.NewEntry(&logrus.Logger{
			Out:       os.Stdout,
			Hooks:     make(logrus.LevelHooks),
			Formatter: localeFormatter,
			Level:     logrus.InfoLevel,
		}),
		Location: localeFormatter.Location,
	}

	for _, option := range options {
		option(logger)
	}

	return logger
}

// OptionalOutput returns an OptionFunc to configure a *Logger to set where log
// messages should output to.
func OptionalOutput(output io.Writer) OptionFunc {
	return func(logger *Logger) {
		logger.Logger.SetOutput(output)
	}
}

// OptionalShardID returns an OptionFunc to configure a *Logger to include a
// shardID field.
func OptionalShardID(shardID int) OptionFunc {
	return func(logger *Logger) {
		logger.Entry = logger.Entry.WithField("shardID", shardID)
	}
}

// OptionalLogLevel returns an OptionFunc to configure a *Logger log level.
func OptionalLogLevel(logLevel string) OptionFunc {
	return func(logger *Logger) {
		logger.UpdateLevel(logLevel)
		logger.UpdateDiscordrus()
	}
}

// OptionalTimezoneLocation returns an OptionFunc to configure a *Logger
// timezone location.
func OptionalTimezoneLocation(timezoneLocation string) OptionFunc {
	return func(logger *Logger) {
		logger.Location = parseTimezoneLocation(logger, timezoneLocation)

		logger.Entry.Logger.Formatter = &locale{
			Location:  logger.Location,
			Formatter: &logrus.TextFormatter{},
		}
	}
}

// OptionalDiscordrus returns an OptionFunc to configure a *Logger to use a
// Discordrus webhook URL.
func OptionalDiscordrus(webhookURL string) OptionFunc {
	return func(logger *Logger) {
		logger.DiscordrusWebHookURL = webhookURL
		logger.UpdateDiscordrus()
	}
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (logger *Logger) WrappedLogger() *logrus.Logger {
	logger.Mutex.Lock()
	defer logger.Mutex.Unlock()

	return logger.Logger
}

// UpdateLevel allows for runtime updates of the logging level.
func (logger *Logger) UpdateLevel(level string) {
	logger.Mutex.Lock()
	defer logger.Mutex.Unlock()

	logger.Logger.SetLevel(parseLevel(level))
}

// UpdateDiscordrus updates the Discordrus integration to use the *Logger
// configuration.
func (logger *Logger) UpdateDiscordrus() {
	logger.Mutex.Lock()
	defer logger.Mutex.Unlock()

	if logger.DiscordrusWebHookURL == "" {
		logger.Logger.Hooks = make(logrus.LevelHooks)
		return
	}

	logger.Logger.Hooks = make(logrus.LevelHooks)

	logger.Logger.AddHook(
		discordrus.NewHook(
			logger.DiscordrusWebHookURL,
			logger.Logger.Level,
			&discordrus.Opts{
				EnableCustomColors: true,
				CustomLevelColors: &discordrus.LevelColors{
					Debug: DebugColor,
					Info:  InfoColor,
					Warn:  WarningColor,
					Error: ErrorColor,
					Panic: PanicColor,
					Fatal: FatalColor,
				},
				TimestampFormat: "Jan 2 15:04:05.00000 MST",
				TimestampLocale: logger.Location,
			},
		),
	)
}

// DiscordGoLogf is an adapter for plugging into DiscordGo's logging system.
func (logger *Logger) DiscordGoLogf(discordgoLevel, caller int, format string, arguments ...interface{}) {
	logger.Mutex.Lock()
	defer logger.Mutex.Unlock()

	switch discordgoLevel {
	case discordgo.LogError:
		logger.Debugf(format, arguments...)
	case discordgo.LogWarning:
		logger.Debugf(format, arguments...)
	case discordgo.LogInformational:
		logger.Debugf(format, arguments...)
	case discordgo.LogDebug:
		logger.Debugf(format, arguments...)
	}
}

type locale struct {
	*time.Location
	logrus.Formatter
}

// Format satisfies the logrus.Formatter interface.
func (locale *locale) Format(log *logrus.Entry) ([]byte, error) {
	if locale.Location == nil {
		return locale.Formatter.Format(log)
	}

	log.Time = log.Time.In(locale.Location)

	return locale.Formatter.Format(log)
}

func parseTimezoneLocation(logger *Logger, location string) *time.Location {
	timezoneLocation, err := time.LoadLocation(location)
	if err != nil {
		logger.WithError(err).Warnf("Error parsing timezone location: %s", location)
		return time.UTC
	}

	return timezoneLocation
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
