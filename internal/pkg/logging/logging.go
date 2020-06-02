// Package logging provides a logrus logging implementation. Configuration
// is determined via environment variables upon startup and logging level may
// be changed at runtime.
package logging

import (
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
	DiscordGof(discordgoLevel, caller int, format string, arguments ...interface{})
}

// Logger wraps a *logrus.Logger instance and provides custom methods.
type Logger struct {
	sync.Mutex
	*logrus.Entry
	Location             *time.Location
	DiscordrusWebHookURL string
}

// New returns a new *Logger instance.
func New(shardID int, logLevel, timezoneLocation, discordrusWebHookURL string) *Logger {
	location := parseTimezoneLocation(timezoneLocation)

	logrusLogger := &logrus.Logger{
		Formatter: &locale{
			&logrus.TextFormatter{},
			location,
		},
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Hooks: make(logrus.LevelHooks),
	}

	log := &Logger{
		Entry:                logrus.NewEntry(logrusLogger).WithField("shardID", shardID),
		Location:             location,
		DiscordrusWebHookURL: discordrusWebHookURL,
	}

	log.UpdateLevel(logLevel)

	return log
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

// UpdateLevel allows for runtime updates of the logging level.
func (log *Logger) UpdateLevel(level string) {
	log.Logger.SetLevel(parseLevel(level))
	log.discordrusIntegration()
}

// DiscordGof is an adaptor for plugging into DiscordGo's logging system.
func (log *Logger) DiscordGof(discordgoLevel, caller int, format string, arguments ...interface{}) {
	switch discordgoLevel {
	case discordgo.LogError:
		log.Errorf(format, arguments...)
	case discordgo.LogWarning:
		log.Warnf(format, arguments...)
	case discordgo.LogInformational:
		log.Infof(format, arguments...)
	case discordgo.LogDebug:
		log.Debugf(format, arguments...)
	}
}

func (log *Logger) discordrusIntegration() {
	log.Lock()
	defer log.Unlock()

	if log.DiscordrusWebHookURL == "" {
		log.Logger.Hooks = make(logrus.LevelHooks)
		return
	}

	log.Logger.Hooks = make(logrus.LevelHooks)

	timeString := time.Now().In(log.Location).String()
	timeZoneToken := strings.Split(timeString, " ")[3]

	log.Logger.AddHook(
		discordrus.NewHook(
			log.DiscordrusWebHookURL,
			log.Logger.Level,
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

type locale struct {
	logrus.Formatter
	*time.Location
}

// Format satisfies the logrus.Formatter interface.
func (locale *locale) Format(log *logrus.Entry) ([]byte, error) {
	log.Time = log.Time.In(locale.Location)

	return locale.Formatter.Format(log)
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
