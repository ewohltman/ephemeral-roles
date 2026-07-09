package operations_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const newRoleID = mock.TestRole

type sessionFunc func() *bot.Client

type roleForMemberTestCase struct {
	name       string
	guildID    snowflake.ID
	userID     snowflake.ID
	roleID     snowflake.ID
	getSession sessionFunc
	testFunc   func(
		t *testing.T,
		getSession sessionFunc,
		guildID, userID, roleID snowflake.ID,
	)
}

const duplicateRequests = 5

func TestNewGateway(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, operations.NewGateway(nil))
}

func TestGateway_CreateRole(t *testing.T) {
	t.Parallel()

	roleNames := []string{mock.TestRoleName, mock.TestRoleName + "2"}

	session, err := mock.NewSession()
	require.NoError(t, err)

	gateway := operations.NewGateway(session)
	waitGroup := &sync.WaitGroup{}

	for _, roleName := range roleNames {
		for range duplicateRequests {
			waitGroup.Go(func() {
				runTestCreateRole(t, gateway, roleName)
			})
		}
	}

	waitGroup.Wait()
}

func TestLookupGuild(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	_, err = operations.LookupGuild(session, mock.TestGuild)
	require.NoError(t, err)

	_, err = operations.LookupGuild(session, mock.TestGuildLarge)
	require.NoError(t, err)
}

func TestAddRoleToMember(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	getSession := func() *bot.Client { return session }

	runRoleForMemberTestCases(t, addRoleToMemberTestCases(getSession))
}

func TestRemoveRoleFromMember(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	getSession := func() *bot.Client { return session }

	runRoleForMemberTestCases(t, removeRoleFromMemberTestCases(getSession))
}

func TestIsDeadlineExceeded(t *testing.T) {
	t.Parallel()

	assert.False(t, operations.IsDeadlineExceeded(io.EOF))
	assert.True(t, operations.IsDeadlineExceeded(&callbacks.EventError{
		Kind: callbacks.KindDeadlineExceeded,
		Err:  context.DeadlineExceeded,
	}))
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
			name:     "*rest.Error http.StatusInternalServerError",
			expected: false,
			err:      &rest.Error{Response: &http.Response{StatusCode: http.StatusInternalServerError}},
		},
		{
			name:     "*rest.Error http.StatusForbidden",
			expected: true,
			err:      &rest.Error{Response: &http.Response{StatusCode: http.StatusForbidden}},
		},
		{
			name:     "wrapped *rest.Error http.StatusForbidden",
			expected: true,
			err:      fmt.Errorf("%w", &rest.Error{Response: &http.Response{StatusCode: http.StatusForbidden}}),
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

	maxGuildsResponse := &rest.Error{Code: operations.APIErrorCodeMaxRoles}

	assert.True(t, operations.IsMaxGuildsResponse(maxGuildsResponse))
}

func TestShouldLogDebug(t *testing.T) {
	t.Parallel()

	assert.False(t, operations.ShouldLogDebug(io.EOF))
	assert.True(t, operations.ShouldLogDebug(&callbacks.EventError{
		Kind: callbacks.KindDeadlineExceeded,
		Err:  context.DeadlineExceeded,
	}))
}

func TestBotHasChannelPermission(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	testChannelWithPermission, ok := session.Caches.Channel(mock.TestChannel)
	require.True(t, ok)

	testChannelWithoutPermission, ok := session.Caches.Channel(mock.TestPrivateChannel)
	require.True(t, ok)

	require.NoError(t, operations.BotHasChannelPermission(session, testChannelWithPermission))
	require.Error(t, operations.BotHasChannelPermission(session, testChannelWithoutPermission))
}

func runTestCreateRole(t *testing.T, gateway callbacks.OperationsGateway, roleName string) {
	t.Helper()

	_, err := gateway.CreateRole(mock.TestGuild, roleName, 0)
	require.NoError(t, err)
}

func addRoleToMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "add role user does not have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
			userID:     mock.TestUser,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
		{
			name:       "add role user does have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
			userID:     mock.TestUser,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
	}
}

func removeRoleFromMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "remove role member does have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
			userID:     mock.TestUser,
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
		{
			name:       "remove role member does not have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
			userID:     mock.TestUser,
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
	}
}

func addNewRoleToMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID snowflake.ID,
) {
	t.Helper()
	roleForMember(t, getSession, guildID, userID, roleID, true)
}

func removeRoleFromMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID snowflake.ID,
) {
	t.Helper()
	roleForMember(t, getSession, guildID, userID, roleID, false)
}

func roleForMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID snowflake.ID,
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
