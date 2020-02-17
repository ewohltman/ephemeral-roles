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

// String values to be used in other tests corresponding to the objects created
// in the mock session.
const (
	TestGuild          = "testGuild"
	TestChannel        = "testChannel"
	TestPrivateChannel = "testPrivateChannel"
	TestRole           = "testRole"
	TestUser           = "testUser"
)

const (
	largeMemberCount = 100

	unsupportedMockRequest = "unsupported mock request"
)

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
func (log *Logger) UpdateLevel(level string) {
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

	sessionUser := "mockSession"

	session.State.User = &discordgo.User{
		ID:       sessionUser,
		Username: sessionUser,
		Bot:      true,
	}

	err := buildTestGuild(session)
	if err != nil {
		return nil, err
	}

	err = buildLargeMemberGuild(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func buildTestGuild(session *discordgo.Session) error {
	testGuild, err := addGuild(session, TestGuild)
	if err != nil {
		return err
	}

	err = addChannel(session, testGuild, TestChannel)
	if err != nil {
		return err
	}

	err = addRole(session, testGuild, TestRole)
	if err != nil {
		return err
	}

	err = addMember(session, testGuild, TestUser)
	if err != nil {
		return err
	}

	return nil
}

func buildLargeMemberGuild(session *discordgo.Session) error {
	testGuild, err := addGuild(session, TestGuild+"2")
	if err != nil {
		return err
	}

	for i := 0; i < largeMemberCount; i++ {
		testUser := fmt.Sprintf("%s-%d", TestUser, i)

		err = addMember(session, testGuild, testUser)
		if err != nil {
			return err
		}
	}

	return nil
}

// SessionClose closes a *discordgo.Session instance and if an error is encountered,
// the provided testingInstance logs the error and marks the test as failed.
func SessionClose(testingInstance TestingInstance, session *discordgo.Session) {
	err := session.Close()
	if err != nil {
		testingInstance.Error(err)
	}
}

func addGuild(session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild := &discordgo.Guild{
		ID:   guildID,
		Name: guildID,
	}

	return guild, session.State.GuildAdd(guild)
}

func addChannel(session *discordgo.Session, guild *discordgo.Guild, channelID string) error {
	channel := &discordgo.Channel{
		ID:      channelID,
		Name:    channelID,
		GuildID: guild.ID,
	}

	guild.Channels = append(guild.Channels, channel)

	return session.State.ChannelAdd(channel)
}

func addRole(session *discordgo.Session, guild *discordgo.Guild, roleID string) error {
	role := &discordgo.Role{
		ID:   roleID,
		Name: roleID,
	}

	guild.Roles = append(guild.Roles, role)

	return session.State.RoleAdd(guild.ID, role)
}

func addMember(session *discordgo.Session, guild *discordgo.Guild, userID string) error {
	member := &discordgo.Member{
		GuildID: guild.ID,
		User: &discordgo.User{
			ID:       userID,
			Username: userID,
		},
	}

	guild.Members = append(guild.Members, member)
	guild.MemberCount++

	return session.State.MemberAdd(member)
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

	if channel == TestPrivateChannel {
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
