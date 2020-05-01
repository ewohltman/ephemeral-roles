package mock

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

// Logger is a mock logger to suppress printing any actual log messages.
type Logger struct {
	*logrus.Logger
}

// NewLogger provides mock *Logger instance.
func NewLogger() *Logger {
	log := &Logger{
		&logrus.Logger{
			Formatter: &logrus.TextFormatter{},
			Out:       ioutil.Discard,
			Level:     logrus.InfoLevel,
			Hooks:     make(logrus.LevelHooks),
		},
	}

	return log
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

// UpdateLevel is a mock stub of the logging.Logger UpdateLevel method.
func (log *Logger) UpdateLevel(level string) {
	// Nop
}

// DiscordGof is an adaptor for plugging into DiscordGo's logging system.
func (log *Logger) DiscordGof(discordgoLevel, caller int, format string, arguments ...interface{}) {
	// Nop
}
