package mock_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewLogger(t *testing.T) {
	log := mock.NewLogger()

	if log == nil {
		t.Fatal("unexpected nil Logger")
	}
}

func TestLogger_WrappedLogger(t *testing.T) {
	log := mock.NewLogger().WrappedLogger()

	if log == nil {
		t.Fatal("unexpected nil wrapped *logrus.Logger")
	}
}

func TestLogger_UpdateLevel(_ *testing.T) {
	mock.NewLogger().UpdateLevel("info")
}

func TestLogger_DiscordGoLogf(_ *testing.T) {
	mock.NewLogger().DiscordGoLogf(discordgo.LogDebug, 0, "Test: %d", 123)
}
