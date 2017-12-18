// Package logging should be used by other packages to get a pointer to the
// global logrus logging instance via the Instance() function call
//
// Logging configuration is determined via environment variables upon startup.
// These variables may be manipulated during runtime and then reflected in the
// global logging instance via the Reinitialize() function call
package logging

import (
	"os"
	"strings"

	"github.com/kz/discordrus"
	"github.com/sirupsen/logrus"
)

// log is a global logrus instance pointer
var log *logrus.Logger

func init() {
	Reinitialize()
}

// Instance returns the global logger instance pointer
func Instance() *logrus.Logger {
	return log
}

// Reinitialize will create a new logging instance configured from the
// environment, updating the global pointer to allow any previous logging
// instance to be garbage collected
func Reinitialize() {
	log = logrus.New()
	logLevel := environmentLevel()

	log.Formatter = &logrus.TextFormatter{}
	log.Level = logLevel
	log.Out = os.Stdout
	log.AddHook(
		discordrus.NewHook(
			// Use environment variable for security reasons
			os.Getenv("DISCORDRUS_WEBHOOK_URL"),
			// Set minimum level to DebugLevel to receive all log entries
			log.Level,
			&discordrus.Opts{
				Username:           "",
				Author:             "",                     // Setting this to a non-empty string adds the author text to the message header
				DisableTimestamp:   false,                  // Setting this to true will disable timestamps from appearing in the footer
				TimestampFormat:    "Jan 2 15:04:05.00000", // The timestamp takes this format; if it is unset, it will take logrus' default format
				EnableCustomColors: true,                   // If set to true, the below CustomLevelColors will apply
				CustomLevelColors: &discordrus.LevelColors{
					Debug: 10170623,
					Info:  3581519,
					Warn:  14327864,
					Error: 13631488,
					Panic: 13631488,
					Fatal: 13631488,
				},
				DisableInlineFields: true, // If set to true, fields will not appear in columns ("inline")
			},
		),
	)
}

// environmentLevel parses and sets our logging level from the environment
func environmentLevel() logrus.Level {
	envLevel, found := os.LookupEnv("LOG_LEVEL")
	if !found || envLevel == "" {
		return logrus.InfoLevel // Default to info if not defined
	}

	switch strings.ToLower(strings.TrimSpace(envLevel)) {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	}

	return logrus.InfoLevel // Default to info if we cannot parse
}
