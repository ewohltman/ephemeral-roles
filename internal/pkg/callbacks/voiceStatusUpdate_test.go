package callbacks_test

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

func TestHandler_VoiceStateUpdate(t *testing.T) {
	jaegerTracer, jaegerCloser, err := tracer.New("test")
	if err != nil {
		t.Fatalf("Error creating Jaeger tracer: %s", err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			t.Errorf("Error closing Jaeger tracer: %s", err)
		}
	}()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	session.Client = http.NewClient(http.WrapTransport(session.Client.Transport))

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:                     log,
		BotName:                 "testBot",
		BotKeyword:              "testKeyword",
		RolePrefix:              "{eph}",
		JaegerTracer:            jaegerTracer,
		ContextTimeout:          time.Second,
		VoiceStateUpdateCounter: monitor.VoiceStateUpdateCounter(&monitor.Config{Log: log}),
		OperationsGateway:       operations.NewGateway(session),
	}

	sendUpdate(session, handler, mock.TestGuild, "unknownUser", mock.TestChannel)
	sendUpdate(session, handler, mock.TestGuild, mock.TestUser, mock.TestPrivateChannel)
	sendUpdate(session, handler, mock.TestGuild, mock.TestUser, mock.TestChannel)
	sendUpdate(session, handler, mock.TestGuild, mock.TestUser, "")
	sendUpdate(session, handler, mock.TestGuild, mock.TestUser, mock.TestChannel2)
	sendUpdate(session, handler, mock.TestGuild, mock.TestUser, mock.TestChannel)
	sendUpdate(session, handler, mock.TestGuild, mock.TestUser, "")
	sendUpdate(session, handler, mock.TestGuildLarge, mock.TestUser, mock.TestChannel)
	sendUpdate(session, handler, mock.TestGuildLarge, mock.TestUser, "")
}

func sendUpdate(session *discordgo.Session, handler *callbacks.Handler, guildID, userID, channelID string) {
	handler.VoiceStateUpdate(session, &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    userID,
			GuildID:   guildID,
			ChannelID: channelID,
		},
	})
}
