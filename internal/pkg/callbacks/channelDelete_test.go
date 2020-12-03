package callbacks_test

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestHandler_ChannelDelete(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:            log,
		BotName:        "testBot",
		BotKeyword:     "testKeyword",
		RolePrefix:     "{eph}",
		ContextTimeout: time.Second,
	}

	guild := session.State.Guilds[0]
	channel := guild.Channels[0]

	if !foundRole(handler, guild, channel) {
		t.Fatalf("Unable to find ephemeral role for channel %s", channel.Name)
	}

	handler.ChannelDelete(session, &discordgo.ChannelDelete{Channel: channel})

	if foundRole(handler, guild, channel) {
		t.Fatalf("Ephemeral role remains for channel %s", channel.Name)
	}
}

func foundRole(handler *callbacks.Handler, guild *discordgo.Guild, channel *discordgo.Channel) bool {
	for _, guildRole := range guild.Roles {
		if guildRole.Name == handler.RoleNameFromChannel(channel.Name) {
			return true
		}
	}

	return false
}
