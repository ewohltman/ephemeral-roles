package callbacks

import (
	"errors"
	"fmt"
	"testing"
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

func TestMemberNotFound_UnWrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	mnf := &memberNotFound{err: wrappedErr}

	unwrappedErr := mnf.UnWrap()

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

func TestChannelNotFound_UnWrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	cnf := &channelNotFound{err: wrappedErr}

	unwrappedErr := cnf.UnWrap()

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

func TestInsufficientPermission_Is(t *testing.T) {
	inp := &insufficientPermission{}

	if errors.Is(nil, &insufficientPermission{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &insufficientPermission{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(inp, &insufficientPermission{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestInsufficientPermission_UnWrap(t *testing.T) {
	inp := &insufficientPermission{}

	unwrappedErr := inp.UnWrap()

	if unwrappedErr != nil {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: nil",
			unwrappedErr,
		)
	}
}

func TestInsufficientPermission_Error(t *testing.T) {
	inp := &insufficientPermission{}
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
