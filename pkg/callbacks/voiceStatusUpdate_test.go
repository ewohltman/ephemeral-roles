package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"
)

func TestConfig_VoiceStateUpdate(t *testing.T) {
	session, err := mock.Session()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

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

	sendUpdate(session, config, "testChannel")
	sendUpdate(session, config, "")
	sendUpdate(session, config, "testChannel2")
	sendUpdate(session, config, "testChannel")
	sendUpdate(session, config, "")
}

func sendUpdate(s *discordgo.Session, config *Config, channelID string) {
	config.VoiceStateUpdate(
		s,
		&discordgo.VoiceStateUpdate{
			VoiceState: &discordgo.VoiceState{
				UserID:    "testUser",
				GuildID:   "testGuild",
				ChannelID: channelID,
			},
		},
	)
}
