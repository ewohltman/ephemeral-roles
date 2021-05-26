package mock_test

import (
	"testing"

	"github.com/ewohltman/discordgo-mock/mockconstants"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewSession(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	_, err = session.User(mockconstants.TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Guild(mockconstants.TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildMember(mockconstants.TestGuild, mockconstants.TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoles(mockconstants.TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Channel(mockconstants.TestChannel)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoleCreate(mockconstants.TestGuild)
	if err != nil {
		t.Error(err)
	}
}
