package callbacks_test

import (
	"context"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

func TestHandler_VoiceStateUpdate(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:                     log,
		RolePrefix:              rolePrefix,
		VoiceStateUpdateCounter: monitor.VoiceStateUpdateCounter(&monitor.Config{Log: log}),
		OperationsGateway:       operations.NewGateway(session),
	}

	testUserMember, ok := session.Caches.Member(mock.TestGuild, mock.TestUser)
	require.True(t, ok)

	unknownMember := discord.Member{
		GuildID: mock.TestGuild,
		User:    discord.User{ID: snowflake.ID(999999), Username: "unknownUser"},
	}

	unknownChannel := snowflake.ID(999888)

	type testCase struct {
		name      string
		member    discord.Member
		channelID *snowflake.ID
	}

	testCases := []*testCase{
		{name: "unknown user", member: unknownMember, channelID: new(mock.TestChannel)},
		{name: "private channel", member: testUserMember, channelID: new(mock.TestPrivateChannel)},
		{name: "join test channel", member: testUserMember, channelID: new(mock.TestChannel)},
		{name: "join test channel 2", member: testUserMember, channelID: new(mock.TestChannel2)},
		{name: "join unknown channel", member: testUserMember, channelID: new(unknownChannel)},
		{name: "rejoin test channel", member: testUserMember, channelID: new(mock.TestChannel)},
		{name: "disconnect test channel", member: testUserMember, channelID: nil},
	}

	mutex := &sync.Mutex{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sendUpdate(mutex, session, handler, &tc.member, tc.channelID)
		})
	}
}

type errorGateway struct {
	err error
}

func (gateway *errorGateway) CreateRole(snowflake.ID, string, int) (discord.Role, error) {
	return discord.Role{}, gateway.err
}

func TestHandler_VoiceStateUpdate_createRoleErrors(t *testing.T) {
	t.Parallel()

	log := mock.NewLogger()

	testCases := []struct {
		err  error
		name string
	}{
		{
			name: "deadline exceeded",
			err:  context.DeadlineExceeded,
		},
		{
			name: "forbidden",
			err:  &rest.Error{Response: &http.Response{StatusCode: http.StatusForbidden}},
		},
		{
			name: "max number of roles",
			err:  &rest.Error{Code: operations.APIErrorCodeMaxRoles},
		},
		{
			name: "unclassified",
			err:  io.EOF,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			session, err := mock.NewSession()
			require.NoError(t, err)

			handler := &callbacks.Handler{
				Log:                     log,
				RolePrefix:              rolePrefix,
				VoiceStateUpdateCounter: monitor.VoiceStateUpdateCounter(&monitor.Config{Log: log}),
				OperationsGateway:       &errorGateway{err: testCase.err},
			}

			member, ok := session.Caches.Member(mock.TestGuild, mock.TestUser)
			require.True(t, ok)

			// mock.TestChannel2 has no pre-existing ephemeral role, forcing
			// the role-creation path to surface the gateway error.
			channelID := mock.TestChannel2

			sendUpdate(&sync.Mutex{}, session, handler, &member, &channelID)
		})
	}
}

func sendUpdate(
	mutex *sync.Mutex,
	session *bot.Client,
	handler *callbacks.Handler,
	member *discord.Member,
	channelID *snowflake.ID,
) {
	mutex.Lock()
	defer mutex.Unlock()

	handler.VoiceStateUpdate(&events.GuildVoiceStateUpdate{
		GenericGuildVoiceState: &events.GenericGuildVoiceState{
			GenericEvent: events.NewGenericEvent(session, 0, 0),
			VoiceState: discord.VoiceState{
				GuildID:   member.GuildID,
				ChannelID: channelID,
				UserID:    member.User.ID,
			},
			Member: *member,
		},
	})

	// VoiceStateUpdate queues its actual work on the guild's sequencer
	// worker; wait for it so the mutex-serialized test cases above see each
	// other's completed role mutations.
	handler.Flush(member.GuildID)
}
