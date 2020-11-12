package operations_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bwmarrin/discordgo"

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
		ctx context.Context,
		t *testing.T,
		getSession sessionFunc,
		guildID, userID, roleName string,
	)
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

				_, err = operations.LookupGuild(context.Background(), session, mock.TestGuildLarge)
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

func TestCreateRole(t *testing.T) {
	type testCase struct {
		name       string
		guildID    string
		roleName   string
		getSession sessionFunc
		testFunc   func(
			ctx context.Context,
			t *testing.T,
			getSession sessionFunc,
			guildID, roleName string,
		)
	}

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	testCases := []*testCase{
		{
			name:       "create new role",
			guildID:    mock.TestGuild,
			roleName:   newRoleID,
			getSession: func() (*discordgo.Session, error) { return session, nil },
			testFunc:   createNewRole,
		},
		{
			name:       "create existing role",
			guildID:    mock.TestGuild,
			roleName:   newRoleID,
			getSession: func() (*discordgo.Session, error) { return session, nil },
			testFunc:   createNewRole,
		},
	}

	ctx := context.Background()

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			testCase.testFunc(ctx, t, testCase.getSession, testCase.guildID, testCase.roleName)
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

	runRoleForMemberTestCases(context.Background(), t, addRoleToMemberTestCases(getSession))
}

func TestRemoveRoleFromMember(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	getSession := func() (*discordgo.Session, error) { return session, nil }

	runRoleForMemberTestCases(context.Background(), t, removeRoleFromMemberTestCases(getSession))
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

	ctx := context.Background()

	err = operations.BotHasChannelPermission(ctx, session, testChannelWithPermission)
	if err != nil {
		t.Error(err)
	}

	err = operations.BotHasChannelPermission(ctx, session, testChannelWithoutPermission)
	if err == nil {
		t.Error("unexpected nil error")
	}
}

func lookupGuild(t *testing.T, getSession sessionFunc, guildID string) {
	session, err := getSession()
	if err != nil {
		t.Fatal(err)
	}

	guild, err := operations.LookupGuild(context.Background(), session, guildID)
	if err != nil {
		t.Fatal(err)
	}

	if guild.ID != guildID {
		t.Errorf("unexpected guild ID: %s (expected: %q)", guild.ID, guildID)
	}
}

func createNewRole(ctx context.Context, t *testing.T, getSession sessionFunc, guildID, roleID string) {
	session, err := getSession()
	if err != nil {
		t.Fatal(err)
	}

	guild, err := operations.LookupGuild(ctx, session, guildID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = operations.CreateRole(ctx, session, guild, roleID, 0)
	if err != nil {
		t.Errorf("unexpected error creating role: %s", err)
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
	ctx context.Context,
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	roleForMember(ctx, t, getSession, guildID, userID, roleID, true)
}

func removeRoleFromMember(
	ctx context.Context,
	t *testing.T,
	getSession sessionFunc,
	guildID, userID, roleID string,
) {
	roleForMember(ctx, t, getSession, guildID, userID, roleID, false)
}

func roleForMember(
	ctx context.Context,
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
		err = operations.AddRoleToMember(ctx, session, guildID, userID, roleID)
		if err != nil {
			t.Errorf("unexpected error adding role to member: %s", err)
		}
	case false:
		err = operations.RemoveRoleFromMember(ctx, session, guildID, userID, roleID)
		if err != nil {
			t.Errorf("unexpected error removing role from member: %s", err)
		}
	}
}

func runRoleForMemberTestCases(ctx context.Context, t *testing.T, testCases []*roleForMemberTestCase) {
	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			testCase.testFunc(ctx, t, testCase.getSession, testCase.guildID, testCase.userID, testCase.roleID)
		})
	}
}

func isForbiddenResponse(t *testing.T, expected bool, err error) {
	actual := operations.IsForbiddenResponse(err)

	if actual != expected {
		t.Errorf("unexpected forbidden response: %t", actual)
	}
}