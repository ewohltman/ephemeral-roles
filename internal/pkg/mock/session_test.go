package mock_test

import (
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewSession(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	_, err = session.User(mock.TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Guild(mock.TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildMember(mock.TestGuild, mock.TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoles(mock.TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Channel(mock.TestChannel)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoleCreate(mock.TestGuild)
	if err != nil {
		t.Error(err)
	}
}
