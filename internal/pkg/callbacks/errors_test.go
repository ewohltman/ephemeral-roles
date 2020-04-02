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

func TestMemberNotFoundError_Is(t *testing.T) {
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

	if wrappedErr != unwrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestMemberNotFound_Error(t *testing.T) {
	mnf := &memberNotFound{}

	errMsg := mnf.Error()

	if errMsg != memberNotFoundMessage {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			memberNotFoundMessage,
		)
	}

	mnf = &memberNotFound{err: fmt.Errorf(wrapMsg)}

	errMsg = mnf.Error()

	expectedErrMsg := fmt.Sprintf("%s: %s", memberNotFoundMessage, wrapMsg)

	if errMsg != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			expectedErrMsg,
		)
	}
}

func TestChannelNotFoundError_Is(t *testing.T) {
	cnf := &channelNotFound{}

	if errors.Is(nil, &memberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &memberNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(cnf, &memberNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestChannelNotFound_UnWrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	cnf := &channelNotFound{err: wrappedErr}

	unwrappedErr := cnf.UnWrap()

	if wrappedErr != unwrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestChannelNotFound_Error(t *testing.T) {
	cnf := &channelNotFound{}

	errMsg := cnf.Error()

	if errMsg != channelNotFoundMessage {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			channelNotFoundMessage,
		)
	}

	cnf = &channelNotFound{err: fmt.Errorf(wrapMsg)}

	errMsg = cnf.Error()

	expectedErrMsg := fmt.Sprintf("%s: %s", channelNotFoundMessage, wrapMsg)

	if errMsg != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			expectedErrMsg,
		)
	}
}
