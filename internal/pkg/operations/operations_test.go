package operations_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const newRoleID = "newRole"

type sessionFunc func() *discordgo.Session

type roleForMemberTestCase struct {
	name       string
	guildID    string
	userID     string
	roleID     string
	getSession sessionFunc
	testFunc   func(
		t *testing.T,
		getSession sessionFunc,
		guildID, userID, roleName string,
	)
}

const duplicateRequests = 5

func TestNewGateway(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, operations.NewGateway(nil))
}

func TestGateway_Process(t *testing.T) {
	t.Parallel()

	roleNames := []string{mockconstants.TestRole, mockconstants.TestRole + "2"}

	session, err := mock.NewSession()
	require.NoError(t, err)

	gateway := operations.NewGateway(session)
	waitGroup := &sync.WaitGroup{}

	runTestRequestUnknown(t, gateway)

	for _, roleName := range roleNames {
		for range duplicateRequests {
			waitGroup.Go(func() {
				runTestRequestCreateRole(t, gateway, roleName)
			})
		}
	}

	waitGroup.Wait()
}

func TestLookupGuild(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	_, err = operations.LookupGuild(session, mockconstants.TestGuild)
	require.NoError(t, err)

	_, err = operations.LookupGuild(session, mockconstants.TestGuildLarge)
	require.NoError(t, err)
}

func TestAddRoleToMember(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	getSession := func() *discordgo.Session { return session }

	runRoleForMemberTestCases(t, addRoleToMemberTestCases(getSession))
}

func TestRemoveRoleFromMember(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	getSession := func() *discordgo.Session { return session }

	runRoleForMemberTestCases(t, removeRoleFromMemberTestCases(getSession))
}

func TestIsDeadlineExceeded(t *testing.T) {
	t.Parallel()

	assert.False(t, operations.IsDeadlineExceeded(io.EOF))
	assert.True(t, operations.IsDeadlineExceeded(&callbacks.DeadlineExceededError{Err: context.DeadlineExceeded}))
}

func TestIsForbiddenResponse(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		expected bool
		err      error
	}

	testCases := []*testCase{
		{
			name:     "nil error",
			expected: false,
			err:      nil,
		},
		{
			name:     "non-nil error",
			expected: false,
			err:      io.EOF,
		},
		{
			name:     "*discordgo.RESTError http.StatusInternalServerError",
			expected: false,
			err:      &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusInternalServerError}},
		},
		{
			name:     "*discordgo.RESTError http.StatusForbidden",
			expected: true,
			err:      &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusForbidden}},
		},
		{
			name:     "wrapped *discordgo.RESTError http.StatusForbidden",
			expected: true,
			err:      fmt.Errorf("%w", &discordgo.RESTError{Response: &http.Response{StatusCode: http.StatusForbidden}}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, operations.IsForbiddenResponse(testCase.err))
		})
	}
}

func TestIsMaxGuildsResponse(t *testing.T) {
	t.Parallel()

	assert.False(t, operations.IsMaxGuildsResponse(io.EOF))

	maxGuildsResponse := &discordgo.RESTError{
		Response: &http.Response{StatusCode: http.StatusBadRequest},
		Message:  &discordgo.APIErrorMessage{Code: operations.APIErrorCodeMaxRoles},
	}

	assert.True(t, operations.IsMaxGuildsResponse(maxGuildsResponse))
}

func TestShouldLogDebug(t *testing.T) {
	t.Parallel()

	assert.False(t, operations.ShouldLogDebug(io.EOF))
	assert.True(t, operations.ShouldLogDebug(&callbacks.DeadlineExceededError{Err: context.DeadlineExceeded}))
}

func TestBotHasChannelPermission(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	testChannelWithPermission, err := session.State.Channel(mockconstants.TestChannel)
	require.NoError(t, err)

	testChannelWithoutPermission, err := session.State.Channel(mockconstants.TestPrivateChannel)
	require.NoError(t, err)

	err = operations.BotHasChannelPermission(session, testChannelWithPermission)
	require.NoError(t, err)

	err = operations.BotHasChannelPermission(session, testChannelWithoutPermission)
	require.Error(t, err)
}

func runTestRequestUnknown(t *testing.T, gateway callbacks.OperationsGateway) {
	t.Helper()

	requestType := operations.RequestType(-1)

	err := runTest(gateway, &operations.Request{
		Type: requestType,
		CreateRole: &operations.CreateRoleRequest{
			Guild:    &discordgo.Guild{ID: mockconstants.TestGuild},
			RoleName: "",
		},
	})
	require.Error(t, err, "unexpected success for request type %q", requestType)
}

func runTestRequestCreateRole(t *testing.T, gateway callbacks.OperationsGateway, roleName string) {
	t.Helper()

	err := runTest(gateway, &operations.Request{
		Type: operations.CreateRole,
		CreateRole: &operations.CreateRoleRequest{
			Guild:    &discordgo.Guild{ID: mockconstants.TestGuild},
			RoleName: roleName,
		},
	})
	require.NoError(t, err)
}

func runTest(gateway callbacks.OperationsGateway, request *operations.Request) error {
	result := <-gateway.Process(request)

	return result.Err
}

func addRoleToMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "add role user does not have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
		{
			name:       "add role user does have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
	}
}

func removeRoleFromMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "remove role member does have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
		{
			name:       "remove role member does not have",
			guildID:    mockconstants.TestGuild,
			roleID:     newRoleID,
			userID:     mockconstants.TestUser,
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
	}
}

func addNewRoleToMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	t.Helper()

	roleForMember(t, getSession, guildID, userID, roleID, true)
}

func removeRoleFromMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	t.Helper()

	roleForMember(t, getSession, guildID, userID, roleID, false)
}

func roleForMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
	add bool,
) {
	t.Helper()

	session := getSession()

	switch add {
	case true:
		require.NoError(t, operations.AddRoleToMember(session, guildID, userID, roleID))
	case false:
		require.NoError(t, operations.RemoveRoleFromMember(session, guildID, userID, roleID))
	}
}

func runRoleForMemberTestCases(t *testing.T, testCases []*roleForMemberTestCase) {
	t.Helper()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testCase.testFunc(t, testCase.getSession, testCase.guildID, testCase.userID, testCase.roleID)
		})
	}
}
