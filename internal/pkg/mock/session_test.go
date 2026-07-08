package mock_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewSession(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	_, err = session.User(mockconstants.TestUser)
	require.NoError(t, err)

	_, err = session.Guild(mockconstants.TestGuild)
	require.NoError(t, err)

	_, err = session.GuildMember(mockconstants.TestGuild, mockconstants.TestUser)
	require.NoError(t, err)

	_, err = session.GuildRoles(mockconstants.TestGuild)
	require.NoError(t, err)

	_, err = session.Channel(mockconstants.TestChannel)
	require.NoError(t, err)

	_, err = session.GuildRoleCreate(
		mockconstants.TestGuild,
		&discordgo.RoleParams{Name: mockconstants.TestRole},
	)
	require.NoError(t, err)
}
