package callbacks

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	wrapMsg               = "wrapped error"
	invalidErrorAssertion = "Invalid error assertion"
)

func TestMemberNotFound_Is(t *testing.T) {
	mnf := &memberNotFound{}

	if errors.Is(nil, &memberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &memberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(mnf, &memberNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestMemberNotFound_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	mnf := &memberNotFound{err: wrappedErr}

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
	mnf := &memberNotFound{}
	expectedErrMsg := memberNotFoundMessage

	if mnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnf.Error(),
			expectedErrMsg,
		)
	}

	mnf.err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", memberNotFoundMessage, wrapMsg)

	if mnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			mnf.Error(),
			expectedErrMsg,
		)
	}
}

func TestMemberNotFound_Guild(t *testing.T) {
	expected := &discordgo.Guild{Name: mock.TestGuild}
	mnf := &memberNotFound{guild: expected}
	actual := mnf.Guild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_Member(t *testing.T) {
	var expected *discordgo.Member

	mnf := &memberNotFound{}
	actual := mnf.Member()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestMemberNotFound_Channel(t *testing.T) {
	var expected *discordgo.Channel

	mnf := &memberNotFound{}
	actual := mnf.Channel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Is(t *testing.T) {
	cnf := &channelNotFound{}

	if errors.Is(nil, &channelNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &channelNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(cnf, &channelNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestChannelNotFound_Unwrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	cnf := &channelNotFound{err: wrappedErr}

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
	cnf := &channelNotFound{}
	expectedErrMsg := channelNotFoundMessage

	if cnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			cnf.Error(),
			expectedErrMsg,
		)
	}

	cnf.err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", channelNotFoundMessage, wrapMsg)

	if cnf.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			cnf.Error(),
			expectedErrMsg,
		)
	}
}

func TestChannelNotFound_Guild(t *testing.T) {
	expected := &discordgo.Guild{Name: mock.TestGuild}
	cnf := &channelNotFound{guild: expected}
	actual := cnf.Guild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Member(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	cnf := &channelNotFound{member: expected}
	actual := cnf.Member()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestChannelNotFound_Channel(t *testing.T) {
	var expected *discordgo.Channel

	cnf := &channelNotFound{}
	actual := cnf.Channel()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermission_Is(t *testing.T) {
	inp := &insufficientPermissions{}

	if errors.Is(nil, &insufficientPermissions{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &insufficientPermissions{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(inp, &insufficientPermissions{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestInsufficientPermission_Unwrap(t *testing.T) {
	inp := &insufficientPermissions{}

	unwrappedErr := inp.Unwrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestInsufficientPermission_Error(t *testing.T) {
	inp := &insufficientPermissions{}
	expectedErrMsg := insufficientPermissionMessage

	if inp.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			inp.Error(),
			expectedErrMsg,
		)
	}

	inp.err = fmt.Errorf(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", insufficientPermissionMessage, wrapMsg)

	if inp.Error() != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			inp.Error(),
			expectedErrMsg,
		)
	}
}

func TestInsufficientPermissions_Guild(t *testing.T) {
	expected := &discordgo.Guild{Name: mock.TestGuild}
	inp := &insufficientPermissions{guild: expected}
	actual := inp.Guild()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_Member(t *testing.T) {
	expected := &discordgo.Member{User: &discordgo.User{Username: mock.TestUser}}
	inp := &insufficientPermissions{member: expected}
	actual := inp.Member()

	err := deepEqual(actual, expected)
	if err != nil {
		t.Error(err)
	}
}

func TestInsufficientPermissions_Channel(t *testing.T) {
	expected := &discordgo.Channel{Name: mock.TestChannel}
	inp := &insufficientPermissions{channel: expected}
	actual := inp.Channel()

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
