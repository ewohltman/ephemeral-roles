package callbacks_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
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

	guild, err := session.State.Guild(mockconstants.TestGuild)
	require.NoError(t, err)

	channel, err := session.State.Channel(mockconstants.TestChannel)
	require.NoError(t, err)

	require.True(t, foundRole(handler, guild, channel), "Unable to find ephemeral role for channel %s", channel.Name)

	handler.ChannelDelete(session, &discordgo.ChannelDelete{Channel: channel})

	require.False(t, foundRole(handler, guild, channel), "Ephemeral role remains for channel %s", channel.Name)
}

func foundRole(handler *callbacks.Handler, guild *discordgo.Guild, channel *discordgo.Channel) bool {
	ephRoleName := handler.RoleNameFromChannel(channel.Name)

	for _, guildRole := range guild.Roles {
		if guildRole.Name == ephRoleName {
			return true
		}
	}

	return false
}
