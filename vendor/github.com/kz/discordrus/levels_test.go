package discordrus

import (
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestAllLevels ensures that logrus' AllLevels has not changed
func TestAllLevels(t *testing.T) {
	// AllLevels is a slice of all the supported Logrus levels
	var AllLevels = []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}

	if !reflect.DeepEqual(AllLevels, logrus.AllLevels) {
		t.Error("discordrus' AllLevels is not the same as logrus' AllLevels")
	}
}

// TestLevelThreshold ensures that the slice returned contains the correct level
func TestLevelThreshold(t *testing.T) {
	var expectedDebugThreshold = []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
	var expectedPanicThreshold = []logrus.Level{
		logrus.PanicLevel,
	}
	var expectedErrorThreshold = []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}

	// Test extreme boundaries
	debugThreshold := LevelThreshold(logrus.DebugLevel)
	if !reflect.DeepEqual(debugThreshold, expectedDebugThreshold) {
		t.Error("Debug threshold does not match expected slice")
	}

	panicThreshold := LevelThreshold(logrus.PanicLevel)
	if !reflect.DeepEqual(panicThreshold, expectedPanicThreshold) {
		t.Error("Panic threshold does not match expected slice")
	}

	// Test within boundaries
	errorThreshold := LevelThreshold(logrus.ErrorLevel)
	if !reflect.DeepEqual(errorThreshold, expectedErrorThreshold) {
		t.Error("Error threshold does not match expected slice")
	}
}

// TestLevelColor ensures LevelColor is able to return the respect color for both default and custom LevelColors
func TestLevelColor(t *testing.T) {
	// Set up custom LevelColors
	customLevelColors := LevelColors{
		Debug: 1,
		Info:  2,
		Warn:  3,
		Error: 4,
		Panic: 5,
		Fatal: 6,
	}

	// Test default colors
	expectedDefaultColorForError := DefaultLevelColors.Error
	if expectedDefaultColorForError != DefaultLevelColors.LevelColor(logrus.ErrorLevel) {
		t.Error("Error color for default LevelColor is not as expected")
	}

	// Test custom colors
	expectedCustomColorForPanic := customLevelColors.Panic
	if expectedCustomColorForPanic != customLevelColors.LevelColor(logrus.PanicLevel) {
		t.Error("Panic color for custom LevelColor is not as expected")
	}
}
