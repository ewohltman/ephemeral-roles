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

	"github.com/ewohltman/ephemeral-roles/pkg/environment"

	"github.com/kz/discordrus"
	"github.com/sirupsen/logrus"
)

const (
	DebugColor = 10170623
	InfoColor  = 3581519
	WarnColor  = 14327864
	ErrorColor = 13631488
	PanicColor = 13631488
	FatalColor = 13631488
)

type localeFormatter struct {
	logrus.Formatter
	*time.Location
}

// Format satisfies the logrus.Formatter interface.
func (l *localeFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.In(l.Location)

	return l.Formatter.Format(e)
}

// New returns a new *logrus.Logger instance.
func New() *logrus.Logger {
	timestampLocale, err := timeLocalization()
	if err != nil {
		timestampLocale = time.Local
	}

	log := &logrus.Logger{
		Formatter: &localeFormatter{
			&logrus.TextFormatter{},
			timestampLocale,
		},
		Out:   os.Stdout,
		Level: logrus.InfoLevel,
		Hooks: make(logrus.LevelHooks),
	}

	UpdateLevel(log)

	return log
}

// UpdateLevel allows for runtime-updates of the global logging instance's
// level and resets the hooks with new values from the environment.
func UpdateLevel(log *logrus.Logger) {
	// Update our global logging instance log level
	log.SetLevel(environmentLevel())

	// Reset logging hooks
	log.Hooks = make(logrus.LevelHooks)

	// Check/apply `github.com/kz/discordrus` hook integration
	discordrusIntegration(log)
}

// discordrusIntegration checks to see if we can apply an optional integration
// support for a `github.com/kz/discordrus` hook.
func discordrusIntegration(log *logrus.Logger) {
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
						Warn:  WarnColor,
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
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	case "panic":
		logLevel = logrus.PanicLevel
	}

	return logLevel
}
