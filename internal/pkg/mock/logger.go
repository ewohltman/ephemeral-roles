package mock

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is a mock logger to suppress printing any actual log messages.
type Logger struct {
	*logrus.Logger
}

// NewLogger provides mock *Logger instance.
func NewLogger() *Logger {
	log := &Logger{
		Logger: &logrus.Logger{
			// Out:       ioutil.Discard,
			Out:       os.Stdout,
			Hooks:     make(logrus.LevelHooks),
			Formatter: &logrus.TextFormatter{},
			Level:     logrus.DebugLevel,
		},
	}

	return log
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (logger *Logger) WrappedLogger() *logrus.Logger {
	return logger.Logger
}

// UpdateLevel is a mock stub of the *logging.Logger UpdateLevel method.
func (logger *Logger) UpdateLevel(level string) {
	// Nop
}

// UpdateDiscordrus is a mock stub of the *logging.Logger UpdateDiscordrus
// method.
func (logger *Logger) UpdateDiscordrus() {
	// Nop
}

// DiscordGoLogf is a mock stub of the *logging.Logger DiscordGoLogf method.
func (logger *Logger) DiscordGoLogf(discordgoLevel, caller int, format string, arguments ...interface{}) {
	// Nop
}
