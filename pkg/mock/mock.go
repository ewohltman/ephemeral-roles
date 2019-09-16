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
		Username: "testUser",
	}

	session.State.User = testUser

	err := session.State.GuildAdd(
		&discordgo.Guild{
			ID:   "testGuild",
			Name: "testGuild",
			Roles: []*discordgo.Role{
				{
					ID:   "testRole",
					Name: "testRole",
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	err = session.State.ChannelAdd(
		&discordgo.Channel{
			ID:      "testChannel",
			Name:    "testChannel",
			GuildID: "testGuild",
		},
	)
	if err != nil {
		return nil, err
	}

	err = session.State.MemberAdd(
		&discordgo.Member{
			User:    testUser,
			Nick:    "testUser",
			GuildID: "testGuild",
			Roles:   []string{"testRole"},
		},
	)

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
	case strings.Contains(r.URL.Path, "channels"):
		respBody = channelsResponse()
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
    "username": "testUser"
}
`)
}

func guildsResponse(r *http.Request) []byte {
	switch r.Method {
	case http.MethodGet:
		switch {
		case strings.Contains(r.URL.Path, "roles"):
			return guildRolesResponse()
		}
	case http.MethodPost:
		return addGuildRoleResponse()
	}

	return []byte("{}")
}

func guildRolesResponse() []byte {
	return []byte(`
[
    {
        "id": "testRole",
        "name": "testRole"
    }
]
`)
}

func addGuildRoleResponse() []byte {
	return []byte(`
{
    "id": "testRole",
    "name": "testRole"
}
`)
}

func channelsResponse() []byte {
	return []byte(`
{
    "id": "testChannel",
    "name": "testChannel"
}
`)
}
