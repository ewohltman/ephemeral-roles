package callbacks_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
)

const (
	wrapMsg               = "wrapped error"
	invalidErrorAssertion = "Invalid error assertion"
)

func TestRoleNotFound_Error(t *testing.T) {
	t.Parallel()

	rnf := &callbacks.RoleNotFoundError{}

	if rnf.Error() == "" {
		t.Error("unexpected empty error message")
	}
}

func TestMemberNotFound_Is(t *testing.T) {
	t.Parallel()

	mnf := &callbacks.MemberNotFoundError{}

	if errors.Is(nil, &callbacks.MemberNotFoundError{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.MemberNotFoundError{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnf, &callbacks.MemberNotFoundError{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestMemberNotFound_Unwrap(t *testing.T) {
	t.Parallel()

	wrappedErr := fmt.Errorf(wrapMsg)

	mnf := &callbacks.MemberNotFoundError{Err: wrappedErr}

	unwrappedErr := mnf.Unwrap()

	if !errors.Is(unwrappedErr, wrappedErr) {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestMemberNotFound_Error(t *testing.T) {
	t.Parallel()

	mnf := &callbacks.MemberNotFoundError{}
	expectedErrMsg := callbacks.MemberNotFoundMessage

	if mnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnf.Error(),
			expectedErrMsg,
		)
	}

	mnf.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if mnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnf.Error(),
			expectedErrMsg,
		)
	}
}

func TestMemberNotFound_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Guild{Name: mockconstants.TestGuild}
	mnf := &callbacks.MemberNotFoundError{Guild: expected}
	actual := mnf.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_ForMember(t *testing.T) {
	t.Parallel()

	var expected *discordgo.Member

	mnf := &callbacks.MemberNotFoundError{}
	actual := mnf.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_InChannel(t *testing.T) {
	t.Parallel()

	var expected *discordgo.Channel

	mnf := &callbacks.MemberNotFoundError{}
	actual := mnf.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Is(t *testing.T) {
	t.Parallel()

	cnf := &callbacks.ChannelNotFoundError{}

	if errors.Is(nil, &callbacks.ChannelNotFoundError{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.ChannelNotFoundError{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(cnf, &callbacks.ChannelNotFoundError{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestChannelNotFound_Unwrap(t *testing.T) {
	t.Parallel()

	wrappedErr := fmt.Errorf(wrapMsg)

	cnf := &callbacks.ChannelNotFoundError{Err: wrappedErr}

	unwrappedErr := cnf.Unwrap()

	if !errors.Is(unwrappedErr, wrappedErr) {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestChannelNotFound_Error(t *testing.T) {
	t.Parallel()

	cnf := &callbacks.ChannelNotFoundError{}
	expectedErrMsg := callbacks.ChannelNotFoundMessage

	if cnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			cnf.Error(),
			expectedErrMsg,
		)
	}

	cnf.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if cnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			cnf.Error(),
			expectedErrMsg,
		)
	}
}

func TestChannelNotFound_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Guild{Name: mockconstants.TestGuild}
	cnf := &callbacks.ChannelNotFoundError{Guild: expected}
	actual := cnf.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Member{User: &discordgo.User{Username: mockconstants.TestUser}}
	cnf := &callbacks.ChannelNotFoundError{Member: expected}
	actual := cnf.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_InChannel(t *testing.T) {
	t.Parallel()

	var expected *discordgo.Channel

	cnf := &callbacks.ChannelNotFoundError{}
	actual := cnf.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermission_Is(t *testing.T) {
	t.Parallel()

	inp := &callbacks.InsufficientPermissionsError{}

	if errors.Is(nil, &callbacks.InsufficientPermissionsError{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.InsufficientPermissionsError{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(inp, &callbacks.InsufficientPermissionsError{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestInsufficientPermission_Unwrap(t *testing.T) {
	t.Parallel()

	inp := &callbacks.InsufficientPermissionsError{}

	unwrappedErr := inp.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestInsufficientPermission_Error(t *testing.T) {
	t.Parallel()

	inp := &callbacks.InsufficientPermissionsError{}
	expectedErrMsg := callbacks.InsufficientPermissionMessage

	if inp.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			inp.Error(),
			expectedErrMsg,
		)
	}

	inp.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if inp.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			inp.Error(),
			expectedErrMsg,
		)
	}
}

func TestInsufficientPermissions_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Guild{Name: mockconstants.TestGuild}
	inp := &callbacks.InsufficientPermissionsError{Guild: expected}
	actual := inp.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Member{User: &discordgo.User{Username: mockconstants.TestUser}}
	inp := &callbacks.InsufficientPermissionsError{Member: expected}
	actual := inp.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_InChannel(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Channel{Name: mockconstants.TestChannel}
	inp := &callbacks.InsufficientPermissionsError{Channel: expected}
	actual := inp.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxNumberOfRoles_Is(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.MaxNumberOfRolesError{}

	if errors.Is(nil, &callbacks.MaxNumberOfRolesError{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.MaxNumberOfRolesError{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnr, &callbacks.MaxNumberOfRolesError{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestMaxNumberOfRoles_Unwrap(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.MaxNumberOfRolesError{}

	unwrappedErr := mnr.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestMaxNumberOfRoles_Error(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.MaxNumberOfRolesError{}
	expectedErrMsg := callbacks.MaxNumberOfRolesMessage

	if mnr.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnr.Error(),
			expectedErrMsg,
		)
	}

	mnr.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if mnr.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnr.Error(),
			expectedErrMsg,
		)
	}
}

func TestMaxNumberOfRoles_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Guild{Name: mockconstants.TestGuild}
	mnr := &callbacks.MaxNumberOfRolesError{Guild: expected}
	actual := mnr.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxNumberOfRoles_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Member{User: &discordgo.User{Username: mockconstants.TestUser}}
	mnr := &callbacks.MaxNumberOfRolesError{Member: expected}
	actual := mnr.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxNumberOfRoles_InChannel(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Channel{Name: mockconstants.TestChannel}
	mnr := &callbacks.MaxNumberOfRolesError{Channel: expected}
	actual := mnr.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestDeadlineExceeded_Is(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.DeadlineExceededError{}

	if errors.Is(nil, &callbacks.DeadlineExceededError{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.DeadlineExceededError{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnr, &callbacks.DeadlineExceededError{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestDeadlineExceeded_Unwrap(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.DeadlineExceededError{}

	unwrappedErr := mnr.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestDeadlineExceeded_Error(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.DeadlineExceededError{}
	expectedErrMsg := callbacks.DeadlineExceededMessage

	if mnr.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnr.Error(),
			expectedErrMsg,
		)
	}

	mnr.Err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	if mnr.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnr.Error(),
			expectedErrMsg,
		)
	}
}

func TestDeadlineExceeded_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Guild{Name: mockconstants.TestGuild}
	mnr := &callbacks.DeadlineExceededError{Guild: expected}
	actual := mnr.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestDeadlineExceeded_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Member{User: &discordgo.User{Username: mockconstants.TestUser}}
	mnr := &callbacks.DeadlineExceededError{Member: expected}
	actual := mnr.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestDeadlineExceeded_InChannel(t *testing.T) {
	t.Parallel()

	expected := &discordgo.Channel{Name: mockconstants.TestChannel}
	mnr := &callbacks.DeadlineExceededError{Channel: expected}
	actual := mnr.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func deepEqual(actual, expected interface{}) error {
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf(
			"unexpected result. Got: %+v, Expected: %+v",
			actual,
			expected,
		)
	}

	return nil
}
