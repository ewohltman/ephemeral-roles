package discordrus

import (
	"github.com/sirupsen/logrus"
)

// LevelColors is a struct of the possible colors used in Discord color format (0x[RGB] converted to int)
type LevelColors struct {
	Debug int
	Info  int
	Warn  int
	Error int
	Panic int
	Fatal int
}

// DefaultLevelColors is a struct of the default colors used
var DefaultLevelColors = LevelColors{
	Debug: 10170623,
	Info:  3581519,
	Warn:  14327864,
	Error: 13631488,
	Panic: 13631488,
	Fatal: 13631488,
}

// LevelThreshold returns a slice of all the levels above and including the level specified
func LevelThreshold(l logrus.Level) []logrus.Level {
	return logrus.AllLevels[:l+1]
}

// LevelColor returns the respective color for the logrus level
func (lc LevelColors) LevelColor(l logrus.Level) int {
	switch l {
	case logrus.DebugLevel:
		return lc.Debug
	case logrus.InfoLevel:
		return lc.Info
	case logrus.WarnLevel:
		return lc.Warn
	case logrus.ErrorLevel:
		return lc.Fatal
	case logrus.PanicLevel:
		return lc.Panic
	default:
		return lc.Warn
	}
}
