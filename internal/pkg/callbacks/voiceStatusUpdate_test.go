package callbacks_test

import (
	"sync"
	"testing"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

func TestHandler_VoiceStateUpdate(t *testing.T) {
	t.Parallel()

	jaegerTracer, jaegerCloser, err := tracer.New("test")
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, jaegerCloser.Close())
	})

	session, err := mock.NewSession()
	require.NoError(t, err)

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:                     log,
		RolePrefix:              rolePrefix,
		JaegerTracer:            jaegerTracer,
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
}
