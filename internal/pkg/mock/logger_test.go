package mock_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	log := mock.NewLogger()

	require.NotNil(t, log)
}

func TestLogger_WrappedLogger(t *testing.T) {
	t.Parallel()

	log := mock.NewLogger().WrappedLogger()

	require.NotNil(t, log)
}

func TestLogger_UpdateLevel(t *testing.T) {
	t.Parallel()

	mock.NewLogger().UpdateLevel("info")
}

func TestLogger_DiscordGoLogf(t *testing.T) {
	t.Parallel()

	mock.NewLogger().DiscordGoLogf(discordgo.LogDebug, 0, "Test: %d", 123)
}
