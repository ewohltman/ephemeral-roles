package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestConfig_VoiceStateUpdate(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	log := mock.NewLogger()

	monitorConfig := &monitor.Config{
		Log: log,
	}

	config := &Config{
		Log:                     log,
		BotName:                 "testBot",
		BotKeyword:              "testKeyword",
		RolePrefix:              "testRolePrefix",
		ReadyCounter:            nil,
		MessageCreateCounter:    nil,
		VoiceStateUpdateCounter: monitorConfig.VoiceStateUpdateCounter(),
	}

	sendUpdate(session, config, "unknownUser", "testChannel")
	sendUpdate(session, config, "testUser", "privateChannel")
	sendUpdate(session, config, "testUser", "testChannel")
	sendUpdate(session, config, "testUser", "")
	sendUpdate(session, config, "testUser", "testChannel2")
	sendUpdate(session, config, "testUser", "testChannel")
	sendUpdate(session, config, "testUser", "")
}

func sendUpdate(s *discordgo.Session, config *Config, userID, channelID string) {
	config.VoiceStateUpdate(
		s,
		&discordgo.VoiceStateUpdate{
			VoiceState: &discordgo.VoiceState{
				UserID:    userID,
				GuildID:   "testGuild",
				ChannelID: channelID,
			},
		},
	)
}
