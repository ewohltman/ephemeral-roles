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

func TestUserNotFoundError_Is(t *testing.T) {
	unf := &userNotFound{}

	if errors.Is(nil, &userNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &userNotFound{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(unf, &userNotFound{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestUserNotFound_UnWrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	unf := &userNotFound{err: wrappedErr}

	unwrappedErr := unf.UnWrap()

	if wrappedErr != unwrappedErr {
		t.Errorf(
			"Unexpected wrapped error. Got %s, Expected: %s",
			unwrappedErr,
			wrappedErr,
		)
	}
}

func TestUserNotFound_Error(t *testing.T) {
	unf := &userNotFound{}

	errMsg := unf.Error()

	if errMsg != userNotFoundMessage {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			userNotFoundMessage,
		)
	}

	unf = &userNotFound{err: fmt.Errorf(wrapMsg)}

	errMsg = unf.Error()

	expectedErrMsg := fmt.Sprintf("%s: %s", userNotFoundMessage, wrapMsg)

	if errMsg != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			expectedErrMsg,
		)
	}
}
