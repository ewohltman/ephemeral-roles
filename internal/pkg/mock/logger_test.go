package mock

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger()

	if log == nil {
		t.Fatal("unexpected nil Logger")
	}
}

func TestLogger_WrappedLogger(t *testing.T) {
	log := NewLogger().WrappedLogger()

	if log == nil {
		t.Fatal("unexpected nil wrapped *logrus.Logger")
	}
}

func TestLogger_UpdateLevel(t *testing.T) {
	NewLogger().UpdateLevel("info")
}

func TestLogger_DiscordGof(t *testing.T) {
	NewLogger().DiscordGof(discordgo.LogDebug, 0, "Test: %d", 123)
}
