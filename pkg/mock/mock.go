package mock

import (
	"bytes"
	"fmt"
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
				{
					ID:   "testRoleChannel",
					Name: "testRolePrefix testChannel",
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

	fmt.Println(r.Method + ": " + r.URL.String())

	switch {
	case strings.Contains(r.URL.Path, "users"):
		respBody = usersResponse(r)
	case strings.Contains(r.URL.Path, "channels"):
		respBody = channelsResponse(r)
	case strings.Contains(r.URL.Path, "guilds"):
		switch r.Method {
		case http.MethodPost:
			fallthrough
		case http.MethodPatch:
			respBody = roleCreateResponse()
		}
	}

	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewReader(respBody)),
	}, nil
}

func usersResponse(r *http.Request) []byte {
	pathTokens := strings.Split(r.URL.Path, "/")
	user := pathTokens[len(pathTokens)-1]

	resp := fmt.Sprintf(
		`{"id":"%s","username":"%s"}`,
		user,
		user,
	)

	return []byte(resp)
}

func channelsResponse(r *http.Request) []byte {
	pathTokens := strings.Split(r.URL.Path, "/")
	channel := pathTokens[len(pathTokens)-1]

	resp := fmt.Sprintf(`{"id":"%s","name":"%s"}`,
		channel,
		channel,
	)

	return []byte(resp)
}

func roleCreateResponse() []byte {
	return []byte(`{"id":"newRole","name":"newRole"}`)
}
