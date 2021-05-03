package operations_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const newRoleID = "newRole"

type sessionFunc func() (*discordgo.Session, error)

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
	if operations.NewGateway(nil) == nil {
		t.Error("unexpected nil queue")
	}
}

func TestGateway_Process(t *testing.T) {
	roleNames := []string{mock.TestRole, mock.TestRole + "2"}

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	gateway := operations.NewGateway(session)
	waitGroup := &sync.WaitGroup{}

	runTestRequestUnknown(t, gateway)

	for _, roleName := range roleNames {
		roleName := roleName

		for i := 0; i < duplicateRequests; i++ {
			waitGroup.Add(1)

			go func() {
				defer waitGroup.Done()
				runTestRequestCreateRole(t, gateway, roleName)
			}()
		}
	}

	waitGroup.Wait()
}

func TestLookupGuild(t *testing.T) {
	type testCase struct {
		name       string
		guildID    string
		getSession sessionFunc
		testFunc   func(t *testing.T, getSession sessionFunc, guildID string)
	}

	testCases := []*testCase{
		{
			name:       "empty state",
			guildID:    mock.TestGuild,
			getSession: mock.NewSessionEmptyState,
			testFunc:   lookupGuild,
		},
		{
			name:       "small member state",
			guildID:    mock.TestGuild,
			getSession: mock.NewSession,
			testFunc:   lookupGuild,
		},
		{
			name:    "large member state",
			guildID: mock.TestGuildLarge,
			getSession: func() (*discordgo.Session, error) {
				session, err := mock.NewSessionEmptyState()
				if err != nil {
					return nil, err
				}

				_, err = operations.LookupGuild(session, mock.TestGuildLarge)
				if err != nil {
					return nil, err
				}

				return session, nil
			},
			testFunc: lookupGuild,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			testCase.testFunc(t, testCase.getSession, testCase.guildID)
		})
	}
}

func TestAddRoleToMember(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	getSession := func() (*discordgo.Session, error) { return session, nil }

	runRoleForMemberTestCases(t, addRoleToMemberTestCases(getSession))
}

func TestRemoveRoleFromMember(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	getSession := func() (*discordgo.Session, error) { return session, nil }

	runRoleForMemberTestCases(t, removeRoleFromMemberTestCases(getSession))
}

func TestIsDeadlineExceeded(t *testing.T) {
	if operations.IsDeadlineExceeded(io.EOF) {
		t.Errorf("Unexpected success")
	}

	if !operations.IsDeadlineExceeded(&callbacks.DeadlineExceeded{Err: context.DeadlineExceeded}) {
		t.Errorf("Unexpected failure")
	}
}

func TestIsForbiddenResponse(t *testing.T) {
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
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			isForbiddenResponse(t, testCase.expected, testCase.err)
		})
	}
}

func TestIsMaxGuildsResponse(t *testing.T) {
	if operations.IsMaxGuildsResponse(io.EOF) {
		t.Errorf("Unexpected success")
	}

	maxGuildsResponse := &discordgo.RESTError{
		Response: &http.Response{StatusCode: http.StatusBadRequest},
		Message:  &discordgo.APIErrorMessage{Code: operations.APIErrorCodeMaxRoles},
	}

	if !operations.IsMaxGuildsResponse(maxGuildsResponse) {
		t.Errorf("Unexpected failure")
	}
}

func TestShouldLogDebug(t *testing.T) {
	if operations.ShouldLogDebug(io.EOF) {
		t.Errorf("Unexpected success")
	}

	if !operations.ShouldLogDebug(&callbacks.DeadlineExceeded{Err: context.DeadlineExceeded}) {
		t.Errorf("Unexpected failure")
	}
}

func TestBotHasChannelPermission(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	testChannelWithPermission, err := session.State.Channel(mock.TestChannel)
	if err != nil {
		t.Fatal(err)
	}

	testChannelWithoutPermission, err := session.State.Channel(mock.TestPrivateChannel)
	if err != nil {
		t.Fatal(err)
	}

	err = operations.BotHasChannelPermission(session, testChannelWithPermission)
	if err != nil {
		t.Error(err)
	}

	err = operations.BotHasChannelPermission(session, testChannelWithoutPermission)
	if err == nil {
		t.Error("unexpected nil error")
	}
}

func runTestRequestUnknown(t *testing.T, gateway callbacks.OperationsGateway) {
	runTest(t, gateway, true, &operations.Request{
		Type: operations.RequestType(-1),
	})
}

func runTestRequestCreateRole(t *testing.T, gateway callbacks.OperationsGateway, roleName string) {
	runTest(t, gateway, false, &operations.Request{
		Type: operations.CreateRole,
		CreateRole: &operations.CreateRoleRequest{
			Guild:    &discordgo.Guild{ID: mock.TestGuild},
			RoleName: roleName,
		},
	})
}

func runTest(t *testing.T, gateway callbacks.OperationsGateway, expectError bool, request *operations.Request) {
	resultChannel := operations.NewResultChannel()

	gateway.Process(resultChannel, request)

	result := <-resultChannel

	_, resultError := result.(error)
	if resultError != expectError {
		if resultError {
			t.Errorf("%s", result)
			return
		}

		t.Errorf("unexpected success for request type %q", request.Type)
	}
}

func lookupGuild(t *testing.T, getSession sessionFunc, guildID string) {
	session, err := getSession()
	if err != nil {
		t.Fatal(err)
	}

	guild, err := operations.LookupGuild(session, guildID)
	if err != nil {
		t.Fatal(err)
	}

	if guild.ID != guildID {
		t.Errorf("unexpected guild ID: %s (expected: %q)", guild.ID, guildID)
	}
}

func addRoleToMemberTestCases(getSession sessionFunc) []*roleForMemberTestCase {
	return []*roleForMemberTestCase{
		{
			name:       "add role user does not have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
			getSession: getSession,
			testFunc:   addNewRoleToMember,
		},
		{
			name:       "add role user does have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
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
			getSession: getSession,
			testFunc:   removeRoleFromMember,
		},
		{
			name:       "remove role member does not have",
			guildID:    mock.TestGuild,
			roleID:     newRoleID,
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
	roleForMember(t, getSession, guildID, userID, roleID, true)
}

func removeRoleFromMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	roleForMember(t, getSession, guildID, userID, roleID, false)
}

func roleForMember(
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
	add bool,
) {
	session, err := getSession()
	if err != nil {
		t.Fatal(err)
	}

	switch add {
	case true:
		err = operations.AddRoleToMember(session, guildID, userID, roleID)
		if err != nil {
			t.Errorf("unexpected error adding role to member: %s", err)
		}
	case false:
		err = operations.RemoveRoleFromMember(session, guildID, userID, roleID)
		if err != nil {
			t.Errorf("unexpected error removing role from member: %s", err)
		}
	}
}

func runRoleForMemberTestCases(t *testing.T, testCases []*roleForMemberTestCase) {
	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			testCase.testFunc(t, testCase.getSession, testCase.guildID, testCase.userID, testCase.roleID)
		})
	}
}

func isForbiddenResponse(t *testing.T, expected bool, err error) {
	actual := operations.IsForbiddenResponse(err)

	if actual != expected {
		t.Errorf("unexpected forbidden response: %t", actual)
	}
}
