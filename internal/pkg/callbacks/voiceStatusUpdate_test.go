package callbacks_test

import (
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
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

	type testCase struct {
		name      string
		guildID   string
		userID    string
		channelID string
	}

	testCases := []*testCase{
		{
			name:      "unknown user",
			guildID:   mockconstants.TestGuild,
			userID:    "unknownUser",
			channelID: mockconstants.TestChannel,
		},
		{
			name:      "private channel",
			guildID:   mockconstants.TestGuild,
			userID:    mockconstants.TestUser,
			channelID: mockconstants.TestPrivateChannel,
		},
		{
			name:      "join test channel",
			guildID:   mockconstants.TestGuild,
			userID:    mockconstants.TestUser,
			channelID: mockconstants.TestChannel,
		},
		{
			name:      "join test channel 2",
			guildID:   mockconstants.TestGuild,
			userID:    mockconstants.TestUser,
			channelID: mockconstants.TestChannel2,
		},
		{
			name:      "join unknown channel",
			guildID:   mockconstants.TestGuild,
			userID:    mockconstants.TestUser,
			channelID: "unknownChannel",
		},
		{
			name:      "rejoin test channel",
			guildID:   mockconstants.TestGuild,
			userID:    mockconstants.TestUser,
			channelID: mockconstants.TestChannel,
		},
		{
			name:      "disconnect test channel",
			guildID:   mockconstants.TestGuild,
			userID:    mockconstants.TestUser,
			channelID: "",
		},
	}

	mutex := &sync.Mutex{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sendUpdate(
				mutex,
				session,
				handler,
				tc.guildID, tc.userID, tc.channelID,
			)
		})
	}
}

func sendUpdate(
	mutex *sync.Mutex,
	session *discordgo.Session,
	handler *callbacks.Handler,
	guildID, userID, channelID string,
) {
	mutex.Lock()
	defer mutex.Unlock()

	handler.VoiceStateUpdate(session, &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    userID,
			GuildID:   guildID,
			ChannelID: channelID,
		},
	})
}
