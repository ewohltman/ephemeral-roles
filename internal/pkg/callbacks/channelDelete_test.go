package callbacks_test

import (
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/pkg/mockconstants"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestHandler_ChannelDelete(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:            log,
		BotName:        "testBot",
		BotKeyword:     "testKeyword",
		RolePrefix:     "{eph}",
		ContextTimeout: time.Second,
	}

	guild, err := session.State.Guild(mockconstants.TestGuild)
	if err != nil {
		t.Fatal(err)
	}

	channel, err := session.State.Channel(mockconstants.TestChannel)
	if err != nil {
		t.Fatal(err)
	}

	if !foundRole(handler, guild, channel) {
		t.Fatalf("Unable to find ephemeral role for channel %s", channel.Name)
	}

	handler.ChannelDelete(session, &discordgo.ChannelDelete{Channel: channel})

	if foundRole(handler, guild, channel) {
		t.Fatalf("Ephemeral role remains for channel %s", channel.Name)
	}
}

func foundRole(handler *callbacks.Handler, guild *discordgo.Guild, channel *discordgo.Channel) bool {
	ephRoleName := handler.RoleNameFromChannel(channel.Name)

	for _, guildRole := range guild.Roles {
		if guildRole.Name == ephRoleName {
			return true
		}
	}

	return false
}
