package mock

import "testing"

func TestNewSession(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer SessionClose(t, session)

	_, err = session.User(TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Guild(TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildMember(TestGuild, TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoles(TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Channel(TestChannel)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoleCreate(TestGuild)
	if err != nil {
		t.Error(err)
	}
}
