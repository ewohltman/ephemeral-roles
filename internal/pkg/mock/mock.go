// Package mock provides implementations for mocking objects and endpoints for
// unit testing.
package mock

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

const unsupportedMockRequest = "unsupported mock request"

// TestingInstance is an interface intended for testing.T and testing.B
// instances.
type TestingInstance interface {
	Error(args ...interface{})
}

// Logger is a mock logger to suppress printing any actual log messages.
type Logger struct {
	*logrus.Logger
}

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

// UpdateLevel is a mock stub of the logging.Logger UpdateLevel method.
func (log *Logger) UpdateLevel() {
	// Nop
}

// NewLogger provides mock *Logger instance.
func NewLogger() *Logger {
	log := &Logger{
		&logrus.Logger{
			Formatter: &logrus.TextFormatter{},
			Out:       ioutil.Discard,
			Level:     logrus.InfoLevel,
			Hooks:     make(logrus.LevelHooks),
		},
	}

	return log
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

// NewSession provides a *discordgo.Session instance to be used in unit testing
// by mocking out Discord API endpoints.
func NewSession() (*discordgo.Session, error) {
	session := &discordgo.Session{
		State:        discordgo.NewState(),
		StateEnabled: true,
		Ratelimiter:  discordgo.NewRatelimiter(),
		Client:       mockRestClient(),
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
	if err != nil {
		return nil, err
	}

	return session, nil
}

// SessionClose closes a *discordgo.Session instance and if an error is encountered,
// the provided testingInstance logs the error and marks the test as failed.
func SessionClose(testingInstance TestingInstance, session *discordgo.Session) {
	err := session.Close()
	if err != nil {
		testingInstance.Error(err)
	}
}

func mockRestClient() *http.Client {
	return &http.Client{
		Transport:     roundTripFunc(discordAPIResponse),
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
}

func discordAPIResponse(r *http.Request) (*http.Response, error) {
	switch {
	case strings.Contains(r.URL.Path, "roles"):
		return roleCreateResponse(r), nil
	case strings.Contains(r.URL.Path, "channels"):
		return channelsResponse(r), nil
	case strings.Contains(r.URL.Path, "users"):
		return usersResponse(r), nil
	}

	return nil, errors.New(unsupportedMockRequest)
}

func roleCreateResponse(r *http.Request) *http.Response {
	switch r.Method {
	case http.MethodPost, http.MethodPatch:
		respBody := []byte(`{"id":"newRole","name":"newRole"}`)
		return newResponse(http.StatusOK, respBody)
	}

	return newResponse(http.StatusMethodNotAllowed, []byte{})
}

func channelsResponse(r *http.Request) *http.Response {
	pathTokens := strings.Split(r.URL.Path, "/")
	channel := pathTokens[len(pathTokens)-1]

	if channel == "privateChannel" {
		return newResponse(http.StatusForbidden, []byte{})
	}

	respBody := []byte(
		fmt.Sprintf(`{"id":"%s","name":"%s"}`,
			channel,
			channel,
		),
	)

	return newResponse(http.StatusOK, respBody)
}

func usersResponse(r *http.Request) *http.Response {
	pathTokens := strings.Split(r.URL.Path, "/")
	user := pathTokens[len(pathTokens)-1]

	respBody := []byte(
		fmt.Sprintf(
			`{"id":"%s","username":"%s"}`,
			user,
			user,
		),
	)

	return newResponse(http.StatusOK, respBody)
}

func newResponse(status int, respBody []byte) *http.Response {
	return &http.Response{
		Status:     http.StatusText(status),
		StatusCode: status,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewReader(respBody)),
	}
}
