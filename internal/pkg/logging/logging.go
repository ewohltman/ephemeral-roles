// Package logging provides a logrus logging implementation. Configuration
// is determined via environment variables upon startup and logging level may
// be changed at runtime.
package logging

import (
	"os"
	"strings"
	"sync"
	"time"

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
}

// JaegerCompatible defines an interface compatible Jaeger.
type JaegerCompatible interface {
	Error(msg string)
	Infof(msg string, args ...interface{})
}

// Logger wraps a *logrus.Logger instance and provides custom methods.
type Logger struct {
	sync.Mutex
	*logrus.Logger
	Location             *time.Location
	DiscordrusWebHookURL string
}

// JaegerLogger wraps a *Logger and provides methods to satisfy the
// JaegerCompatible interface.
type JaegerLogger struct {
	*Logger
}

// New returns a new *Logger instance.
func New(logLevel, timezoneLocation, discordrusWebHookURL string) *Logger {
	location := parseTimezoneLocation(timezoneLocation)

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
		DiscordrusWebHookURL: discordrusWebHookURL,
	}

	log.UpdateLevel(logLevel)

	return log
}

// Error satisfies the JaegerCompatible interface by delegating to the wrapped
// *Logger Error method.
func (jaegerLogger *JaegerLogger) Error(msg string) {
	jaegerLogger.Logger.Error(msg)
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

// UpdateLevel allows for runtime updates of the logging level.
func (log *Logger) UpdateLevel(level string) {
	log.SetLevel(parseLevel(level))
	log.discordrusIntegration()
}

func (log *Logger) discordrusIntegration() {
	if log.DiscordrusWebHookURL == "" {
		return
	}

	log.Lock()
	defer log.Unlock()

	log.Hooks = make(logrus.LevelHooks)

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

func parseTimezoneLocation(location string) *time.Location {
	timezoneLocation, err := time.LoadLocation(location)
	if err != nil {
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
