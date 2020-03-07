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
	unf := &userNotFoundError{}

	if errors.Is(nil, &userNotFoundError{}) {
		t.Error(invalidErrorAssertion)
	}

	if errors.Is(fmt.Errorf(wrapMsg), &userNotFoundError{}) {
		t.Error(invalidErrorAssertion)
	}

	if !errors.Is(unf, &userNotFoundError{}) {
		t.Errorf(invalidErrorAssertion)
	}
}

func TestUserNotFound_UnWrap(t *testing.T) {
	wrappedErr := fmt.Errorf(wrapMsg)

	unf := &userNotFoundError{err: wrappedErr}

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
	unf := &userNotFoundError{}

	errMsg := unf.Error()

	if errMsg != userNotFoundErrorMessage {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			userNotFoundErrorMessage,
		)
	}

	unf = &userNotFoundError{err: fmt.Errorf(wrapMsg)}

	errMsg = unf.Error()

	expectedErrMsg := fmt.Sprintf("%s: %s", userNotFoundErrorMessage, wrapMsg)

	if errMsg != expectedErrMsg {
		t.Errorf(
			"Unexpected error message. Got %s, Expected: %s",
			errMsg,
			expectedErrMsg,
		)
	}
}
