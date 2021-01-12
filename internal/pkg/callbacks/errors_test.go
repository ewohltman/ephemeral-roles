package callbacks_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	wrapMsg               = "wrapped error"
	invalidErrorAssertion = "Invalid error assertion"
)

func TestRoleNotFound_Error(t *testing.T) {
	rnf := &callbacks.RoleNotFound{}

	if rnf.Error() == "" {
		t.Error("unexpected empty error message")
	}
}

func TestMemberNotFound_Is(t *testing.T) {
	mnf := &callbacks.MemberNotFound{}

	if errors.Is(nil, &callbacks.MemberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.MemberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnf, &callbacks.MemberNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestMemberNotFound_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	mnf := &callbacks.MemberNotFound{Err: wrappedErr}

	unwrappedErr := mnf.Unwrap()

	if unwrappedErr != wrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestMemberNotFound_Error(t *testing.T) {
	mnf := &callbacks.MemberNotFound{}
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
	expected := &discordgo.Guild{Name: mock.TestGuild}
	mnf := &callbacks.MemberNotFound{Guild: expected}
	actual := mnf.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_ForMember(t *testing.T) {
	var expected *discordgo.Member

	mnf := &callbacks.MemberNotFound{}
	actual := mnf.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_InChannel(t *testing.T) {
	var expected *discordgo.Channel

	mnf := &callbacks.MemberNotFound{}
	actual := mnf.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Is(t *testing.T) {
	cnf := &callbacks.ChannelNotFound{}

	if errors.Is(nil, &callbacks.ChannelNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.ChannelNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(cnf, &callbacks.ChannelNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestChannelNotFound_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	cnf := &callbacks.ChannelNotFound{Err: wrappedErr}

	unwrappedErr := cnf.Unwrap()

	if unwrappedErr != wrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestChannelNotFound_Error(t *testing.T) {
	cnf := &callbacks.ChannelNotFound{}
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
	expected := &discordgo.Guild{Name: mock.TestGuild}
	cnf := &callbacks.ChannelNotFound{Guild: expected}
	actual := cnf.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_ForMember(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	cnf := &callbacks.ChannelNotFound{Member: expected}
	actual := cnf.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_InChannel(t *testing.T) {
	var expected *discordgo.Channel

	cnf := &callbacks.ChannelNotFound{}
	actual := cnf.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermission_Is(t *testing.T) {
	inp := &callbacks.InsufficientPermissions{}

	if errors.Is(nil, &callbacks.InsufficientPermissions{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.InsufficientPermissions{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(inp, &callbacks.InsufficientPermissions{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestInsufficientPermission_Unwrap(t *testing.T) {
	inp := &callbacks.InsufficientPermissions{}

	unwrappedErr := inp.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestInsufficientPermission_Error(t *testing.T) {
	inp := &callbacks.InsufficientPermissions{}
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
	expected := &discordgo.Guild{Name: mock.TestGuild}
	inp := &callbacks.InsufficientPermissions{Guild: expected}
	actual := inp.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_ForMember(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	inp := &callbacks.InsufficientPermissions{Member: expected}
	actual := inp.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_InChannel(t *testing.T) {
	expected := &discordgo.Channel{Name: mock.TestChannel}
	inp := &callbacks.InsufficientPermissions{Channel: expected}
	actual := inp.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxNumberOfRoles_Is(t *testing.T) {
	mnr := &callbacks.MaxNumberOfRoles{}

	if errors.Is(nil, &callbacks.MaxNumberOfRoles{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.MaxNumberOfRoles{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnr, &callbacks.MaxNumberOfRoles{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestMaxNumberOfRoles_Unwrap(t *testing.T) {
	mnr := &callbacks.MaxNumberOfRoles{}

	unwrappedErr := mnr.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestMaxNumberOfRoles_Error(t *testing.T) {
	mnr := &callbacks.MaxNumberOfRoles{}
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
	expected := &discordgo.Guild{Name: mock.TestGuild}
	mnr := &callbacks.MaxNumberOfRoles{Guild: expected}
	actual := mnr.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxNumberOfRoles_ForMember(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	mnr := &callbacks.MaxNumberOfRoles{Member: expected}
	actual := mnr.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxNumberOfRoles_InChannel(t *testing.T) {
	expected := &discordgo.Channel{Name: mock.TestChannel}
	mnr := &callbacks.MaxNumberOfRoles{Channel: expected}
	actual := mnr.InChannel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestDeadlineExceeded_Is(t *testing.T) {
	mnr := &callbacks.DeadlineExceeded{}

	if errors.Is(nil, &callbacks.DeadlineExceeded{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &callbacks.DeadlineExceeded{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnr, &callbacks.DeadlineExceeded{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestDeadlineExceeded_Unwrap(t *testing.T) {
	mnr := &callbacks.DeadlineExceeded{}

	unwrappedErr := mnr.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestDeadlineExceeded_Error(t *testing.T) {
	mnr := &callbacks.DeadlineExceeded{}
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
	expected := &discordgo.Guild{Name: mock.TestGuild}
	mnr := &callbacks.DeadlineExceeded{Guild: expected}
	actual := mnr.InGuild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestDeadlineExceeded_ForMember(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	mnr := &callbacks.DeadlineExceeded{Member: expected}
	actual := mnr.ForMember()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestDeadlineExceeded_InChannel(t *testing.T) {
	expected := &discordgo.Channel{Name: mock.TestChannel}
	mnr := &callbacks.DeadlineExceeded{Channel: expected}
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
