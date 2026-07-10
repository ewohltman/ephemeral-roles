package callbacks_test

import (
	"testing"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestHandler_ChannelDelete(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:        log,
		RolePrefix: rolePrefix,
	}

	channel, ok := session.Caches.Channel(mock.TestChannel)
	require.True(t, ok)

	require.True(t, foundRole(session, handler, mock.TestGuild, channel),
		"Unable to find ephemeral role for channel %s", channel.Name())

	handler.ChannelDelete(&events.GuildChannelDelete{
		GenericGuildChannel: &events.GenericGuildChannel{
			GenericEvent: events.NewGenericEvent(session, 0, 0),
			ChannelID:    channel.ID(),
			Channel:      channel,
			GuildID:      mock.TestGuild,
		},
	})

	// ChannelDelete queues its actual work on the guild's sequencer worker.
	handler.Flush(mock.TestGuild)

	require.False(t, foundRole(session, handler, mock.TestGuild, channel),
		"Ephemeral role remains for channel %s", channel.Name())
}

func foundRole(session *bot.Client, handler *callbacks.Handler, guildID snowflake.ID, channel discord.GuildChannel) bool {
	ephRoleName := handler.RoleNameFromChannel(channel.Name())

	for role := range session.Caches.Roles(guildID) {
		if role.Name == ephRoleName {
			return true
		}
	}

	return false
}
