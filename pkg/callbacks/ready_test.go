package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"
	"github.com/sirupsen/logrus"
)

func TestConfig_Ready(t *testing.T) {
	session, err := mock.Session()
	if err != nil {
		t.Fatal(err)
	}

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
		ReadyCounter:            monitorConfig.ReadyCounter(),
		MessageCreateCounter:    nil,
		VoiceStateUpdateCounter: nil,
	}

	config.Ready(
		session,
		&discordgo.Ready{
			Guilds: make([]*discordgo.Guild, 0),
		},
	)
}
