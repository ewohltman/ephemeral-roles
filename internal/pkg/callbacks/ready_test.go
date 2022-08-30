package callbacks_test

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestHandler_Ready(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:          log,
		BotName:      "testBot",
		RolePrefix:   "testRolePrefix",
		ReadyCounter: monitor.ReadyCounter(&monitor.Config{Log: log}),
	}

	handler.Ready(session, &discordgo.Ready{
		Guilds: make([]*discordgo.Guild, 0),
	})
}
