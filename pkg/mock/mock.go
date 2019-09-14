package mock

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func Session() (*discordgo.Session, error) {
	session := &discordgo.Session{
		State:        discordgo.NewState(),
		StateEnabled: true,
		Ratelimiter:  discordgo.NewRatelimiter(),
		Client:       DiscordRestClient(),
	}

	testUser := &discordgo.User{
		ID:       "testUser",
		Username: "Test User",
	}

	session.State.User = testUser

	err := session.State.GuildAdd(
		&discordgo.Guild{
			ID:    "testGuild",
			Roles: make([]*discordgo.Role, 0),
		},
	)
	if err != nil {
		return nil, err
	}

	err = session.State.MemberAdd(
		&discordgo.Member{
			User:    testUser,
			Nick:    "Test User",
			GuildID: "testGuild",
			Roles:   make([]string, 0),
		},
	)

	err = session.State.ChannelAdd(
		&discordgo.Channel{
			ID:      "testChannel",
			Name:    "Channel Name",
			GuildID: "testGuild",
		},
	)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func DiscordRestClient() *http.Client {
	return &http.Client{
		Transport:     roundTripFunc(discordAPIResponse),
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
}

func discordAPIResponse(r *http.Request) (*http.Response, error) {
	respBody := []byte("")

	switch {
	case strings.Contains(r.URL.Path, "users"):
		respBody = usersResponse()
	case strings.Contains(r.URL.Path, "guilds"):
		respBody = guildsResponse(r)
	}

	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewReader(respBody)),
	}, nil
}

func usersResponse() []byte {
	return []byte(`
{
    "id": "testUser",
    "username": "Test User"
}
`)
}

func guildsResponse(r *http.Request) []byte {
	switch {
	case strings.Contains(r.URL.Path, "roles"):
		return guildRolesResponse()
	default:
		return []byte("")
	}
}

func guildRolesResponse() []byte {
	return []byte(`
{
    "id": "testRole",
    "name": "Test Role"
}
`)
}
