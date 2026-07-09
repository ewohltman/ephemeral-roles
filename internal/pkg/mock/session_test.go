package mock_test

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewSession(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	_, ok := session.Caches.SelfUser()
	require.True(t, ok)

	_, ok = session.Caches.Guild(mock.TestGuild)
	require.True(t, ok)

	_, ok = session.Caches.Member(mock.TestGuild, mock.TestUser)
	require.True(t, ok)

	_, ok = session.Caches.Role(mock.TestGuild, mock.TestRole)
	require.True(t, ok)

	_, ok = session.Caches.Channel(mock.TestChannel)
	require.True(t, ok)

	_, err = session.Rest.CreateRole(
		mock.TestGuild,
		discord.RoleCreate{Name: mock.TestRoleName},
	)
	require.NoError(t, err)
}
