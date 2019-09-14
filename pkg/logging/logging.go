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
)

// localeFormatter is a custom formatter for logrus
type localeFormatter struct {
	logrus.Formatter
	*time.Location
}

// (*localeFormatter) Format satisfies the logrus.Formatter interface
func (l *localeFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.In(l.Location)

	return l.Formatter.Format(e)
}

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

	// Check/apply `github.com/kz/discordrus` hook integration
	discordrusIntegration(log)

	return log
}

// UpdateLevel allows for runtime-updates of the global logging instance's
// level and resets the hooks with new values from the environment
func UpdateLevel(log *logrus.Logger) {
	// Update our global logging instance log level
	log.SetLevel(environmentLevel())

	// Reset logging hooks
	log.Hooks = make(logrus.LevelHooks)

	// Check/apply `github.com/kz/discordrus` hook integration
	discordrusIntegration(log)
}

// discordrusIntegration checks to see if we can apply an optional integration
// support for a `github.com/kz/discordrus` hook
func discordrusIntegration(log *logrus.Logger) {
	if hookURLString, found := os.LookupEnv("DISCORDRUS_WEBHOOK_URL"); found {
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
						Debug: 10170623,
						Info:  3581519,
						Warn:  14327864,
						Error: 13631488,
						Panic: 13631488,
						Fatal: 13631488,
					},
					DisableTimestamp: false,           // Setting this to true will disable timestamps from appearing in the footer
					TimestampFormat:  timeStampFormat, // The timestamp takes this format; if it is unset, it will take logrus' default format
					TimestampLocale:  timestampLocale, // The timestamp takes it's timezone from the provided locale
				},
			),
		)
	}
}

// timeLocalization returns the *time.Location defined in the environment by
// LOG_TIMEZONE_LOCATION, else defaults to time.Local
func timeLocalization() (timeLocalization *time.Location, err error) {
	envLocation, found := os.LookupEnv("LOG_TIMEZONE_LOCATION")
	if !found {
		err = fmt.Errorf("LOG_TIMEZONE_LOCATION not defined in environment variables")

		return
	}

	if envLocation == "" {
		envLocation = time.Local.String()
	}

	parsedLocation, parseErr := time.LoadLocation(envLocation)
	if parseErr != nil {
		err = fmt.Errorf("unable to parse LOG_TIMEZONE_LOCATION: %s", err.Error())

		return
	}

	timeLocalization = parsedLocation

	return
}

// environmentLevel parses and returns our logging level from the environment
func environmentLevel() (logLevel logrus.Level) {
	logLevel = logrus.InfoLevel // Default to InfoLevel

	envLevel, found := os.LookupEnv("LOG_LEVEL")
	if !found || envLevel == "" {
		return
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

	return
}
