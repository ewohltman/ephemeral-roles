// Package mock provides implementations for mocking objects and endpoints for
// unit testing.
package mock

import (
	"bytes"
	"encoding/json"
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

// RoundTripperFunc allows functions to satisfy the http.RoundTripper
// interface.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface.
func (rt RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// Logger is a mock logger to suppress printing any actual log messages.
type Logger struct {
	*logrus.Logger
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

// WrappedLogger returns the wrapped *logrus.Logger instance.
func (log *Logger) WrappedLogger() *logrus.Logger {
	return log.Logger
}

// UpdateLevel is a mock stub of the logging.Logger UpdateLevel method.
func (log *Logger) UpdateLevel(level string) {
	// Nop
}

// NewMirrorRoundTripper returns an http.RoundTripper that mirrors the request
// body in the response body.
func NewMirrorRoundTripper() http.RoundTripper {
	return RoundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Request:    req,
			}

			if req.Body == nil {
				resp.ContentLength = 0
				resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))

				return resp, nil
			}

			reqBody, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			err = req.Body.Close()
			if err != nil {
				return nil, err
			}

			resp.ContentLength = int64(len(reqBody))
			resp.Body = ioutil.NopCloser(bytes.NewReader(reqBody))

			return resp, nil
		},
	)
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

// SessionClose closes a *discordgo.Session instance and if an error is encountered,
// the provided testingInstance logs the error and marks the test as failed.
func SessionClose(testingInstance TestingInstance, session *discordgo.Session) {
	err := session.Close()
	if err != nil {
		testingInstance.Error(err)
	}
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

	err = addRoles(session, testGuild, mockRoles())
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

func addRoles(session *discordgo.Session, guild *discordgo.Guild, roles discordgo.Roles) error {
	guild.Roles = append(guild.Roles, roles...)

	for _, role := range guild.Roles {
		err := session.State.RoleAdd(guild.ID, role)
		if err != nil {
			return err
		}
	}

	return nil
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
	return &http.Client{Transport: RoundTripperFunc(discordAPIResponse)}
}

func discordAPIResponse(r *http.Request) (*http.Response, error) {
	switch {
	case strings.Contains(r.URL.Path, "users"):
		return usersResponse(r), nil
	case strings.Contains(r.URL.Path, "members"):
		return membersResponse(r), nil
	case strings.Contains(r.URL.Path, "roles"):
		return rolesResponse(r), nil
	case strings.Contains(r.URL.Path, "channels"):
		return channelsResponse(r), nil
	case strings.Contains(r.URL.Path, "guilds"):
		return guildsResponse(r), nil
	}

	return nil, fmt.Errorf(unsupportedMockRequest)
}

func usersResponse(r *http.Request) *http.Response {
	pathTokens := strings.Split(r.URL.Path, "/")
	userID := pathTokens[len(pathTokens)-1]

	respBody, err := json.Marshal(mockUser(userID))
	if err != nil {
		return newResponse(http.StatusInternalServerError, []byte(err.Error()))
	}

	return newResponse(http.StatusOK, respBody)
}

func membersResponse(r *http.Request) *http.Response {
	pathTokens := strings.Split(r.URL.Path, "/")
	userID := pathTokens[len(pathTokens)-1]

	respBody, err := json.Marshal(mockMember(userID))
	if err != nil {
		return newResponse(http.StatusInternalServerError, []byte(err.Error()))
	}

	return newResponse(http.StatusOK, respBody)
}

func rolesResponse(r *http.Request) *http.Response {
	switch r.Method {
	case http.MethodGet:
		respBody, err := json.Marshal(mockRoles())
		if err != nil {
			return newResponse(http.StatusInternalServerError, []byte(err.Error()))
		}

		return newResponse(http.StatusOK, respBody)
	case http.MethodPost:
		respBody, err := json.Marshal(mockRole(TestRole))
		if err != nil {
			return newResponse(http.StatusInternalServerError, []byte(err.Error()))
		}

		return newResponse(http.StatusOK, respBody)
	case http.MethodPatch:
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return newResponse(http.StatusInternalServerError, []byte(err.Error()))
		}

		err = r.Body.Close()
		if err != nil {
			return newResponse(http.StatusInternalServerError, []byte(err.Error()))
		}

		return newResponse(http.StatusOK, reqBody)
	}

	return newResponse(http.StatusMethodNotAllowed, []byte{})
}

func channelsResponse(r *http.Request) *http.Response {
	var (
		respBody []byte
		err      error
	)

	if strings.Contains(r.URL.Path, "guilds") {
		respBody, err = json.Marshal(mockChannels())
	} else {
		respBody, err = json.Marshal(mockChannel(TestChannel))
	}

	if err != nil {
		return newResponse(http.StatusInternalServerError, []byte(err.Error()))
	}

	return newResponse(http.StatusOK, respBody)
}

func guildsResponse(r *http.Request) *http.Response {
	pathTokens := strings.Split(r.URL.Path, "/")
	guildID := pathTokens[len(pathTokens)-1]

	respBody, err := json.Marshal(mockGuild(guildID))
	if err != nil {
		return newResponse(http.StatusInternalServerError, []byte(err.Error()))
	}

	return newResponse(http.StatusOK, respBody)
}

func mockUser(userID string) *discordgo.User {
	return &discordgo.User{
		ID:       userID,
		Username: userID,
	}
}

func mockMember(userID string) *discordgo.Member {
	return &discordgo.Member{
		GuildID: TestGuild,
		User:    mockUser(userID),
		Roles:   mockRoleIDs(mockRoles()),
	}
}

func mockMembers() []*discordgo.Member {
	return []*discordgo.Member{mockMember(TestUser)}
}

func mockRole(roleID string) *discordgo.Role {
	return &discordgo.Role{
		ID:   roleID,
		Name: roleID,
	}
}

func mockRoles() discordgo.Roles {
	return discordgo.Roles{
		mockRole(TestRole),
		mockRole("{eph} " + TestChannel),
	}
}

func mockRoleIDs(roles discordgo.Roles) []string {
	roleIDs := make([]string, len(roles))

	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	return roleIDs
}

func mockChannel(channelID string) *discordgo.Channel {
	return &discordgo.Channel{
		ID:      channelID,
		Name:    channelID,
		GuildID: TestGuild,
	}
}

func mockChannels() []*discordgo.Channel {
	return []*discordgo.Channel{mockChannel(TestChannel)}
}

func mockGuild(guildID string) *discordgo.Guild {
	return &discordgo.Guild{
		ID:       guildID,
		Name:     guildID,
		Roles:    mockRoles(),
		Members:  mockMembers(),
		Channels: mockChannels(),
	}
}

func newResponse(status int, respBody []byte) *http.Response {
	return &http.Response{
		Status:     http.StatusText(status),
		StatusCode: status,
		Header:     make(http.Header),
		Body:       ioutil.NopCloser(bytes.NewReader(respBody)),
	}
}
