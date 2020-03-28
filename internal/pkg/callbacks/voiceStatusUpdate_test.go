package callbacks

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

func TestConfig_VoiceStateUpdate(t *testing.T) {
	jaegerTracer, jaegerCloser, err := tracer.New(mock.NewLogger(), "test")
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

	session.Client = http.NewClient(
		session.Client.Transport,
		jaegerTracer,
		"test-0",
	)

	defer mock.SessionClose(t, session)

	log := mock.NewLogger()
	log.Out = os.Stdout

	monitorConfig := &monitor.Config{
		Log: log,
	}

	config := &Config{
		Log:                     log,
		BotName:                 "testBot",
		BotKeyword:              "testKeyword",
		RolePrefix:              "{eph}",
		ReadyCounter:            nil,
		MessageCreateCounter:    nil,
		VoiceStateUpdateCounter: monitorConfig.VoiceStateUpdateCounter(),
	}

	sendUpdate(session, config, "unknownUser", mock.TestChannel)
	sendUpdate(session, config, mock.TestUser, mock.TestPrivateChannel)
	sendUpdate(session, config, mock.TestUser, mock.TestChannel)
	sendUpdate(session, config, mock.TestUser, "")
	sendUpdate(session, config, mock.TestUser, mock.TestChannel+"x")
	sendUpdate(session, config, mock.TestUser, mock.TestChannel)
	sendUpdate(session, config, mock.TestUser, "")
}

func sendUpdate(session *discordgo.Session, config *Config, userID, channelID string) {
	config.VoiceStateUpdate(
		session,
		&discordgo.VoiceStateUpdate{
			VoiceState: &discordgo.VoiceState{
				UserID:    userID,
				GuildID:   mock.TestGuild,
				ChannelID: channelID,
			},
		},
	)
}
