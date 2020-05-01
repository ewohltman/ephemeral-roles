package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestConfig_Ready(t *testing.T) {
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
		Log:          log,
		BotName:      "testBot",
		BotKeyword:   "testKeyword",
		RolePrefix:   "testRolePrefix",
		ReadyCounter: monitorConfig.ReadyCounter(),
	}

	config.Ready(
		session,
		&discordgo.Ready{
			Guilds: make([]*discordgo.Guild, 0),
		},
	)
}
