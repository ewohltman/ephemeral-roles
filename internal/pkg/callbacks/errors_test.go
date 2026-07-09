package callbacks_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const wrapMsg = "wrapped error"

func TestErrorKind_Message(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		expected string
		kind     callbacks.ErrorKind
	}{
		{expected: callbacks.MemberNotFoundMessage, kind: callbacks.KindMemberNotFound},
		{expected: callbacks.ChannelNotFoundMessage, kind: callbacks.KindChannelNotFound},
		{expected: callbacks.InsufficientPermissionMessage, kind: callbacks.KindInsufficientPermissions},
		{expected: callbacks.MaxNumberOfRolesMessage, kind: callbacks.KindMaxNumberOfRoles},
		{expected: callbacks.DeadlineExceededMessage, kind: callbacks.KindDeadlineExceeded},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, testCase.kind.Message())
	}

	assert.NotEmpty(t, callbacks.ErrorKind(-1).Message())
}

func TestEventError_Error(t *testing.T) {
	t.Parallel()

	eventErr := &callbacks.EventError{Kind: callbacks.KindMemberNotFound}

	require.EqualError(t, eventErr, callbacks.MemberNotFoundMessage)

	eventErr.Err = errors.New(wrapMsg)

	require.EqualError(t, eventErr, fmt.Sprintf("%s: %s", callbacks.MemberNotFoundMessage, wrapMsg))
}

func TestEventError_Unwrap(t *testing.T) {
	t.Parallel()

	require.NoError(t, (&callbacks.EventError{}).Unwrap())

	wrappedErr := errors.New(wrapMsg)
	eventErr := &callbacks.EventError{Kind: callbacks.KindChannelNotFound, Err: wrappedErr}

	require.ErrorIs(t, eventErr, wrappedErr)
}

func TestEventError_As(t *testing.T) {
	t.Parallel()

	eventErr := &callbacks.EventError{
		Kind:    callbacks.KindInsufficientPermissions,
		Guild:   &discord.Guild{Name: mock.TestGuildName},
		Member:  &discord.Member{User: discord.User{Username: mock.TestUserName}},
		Channel: discord.GuildVoiceChannel{},
	}

	wrapped := fmt.Errorf("outer: %w", eventErr)

	unwrapped, ok := errors.AsType[*callbacks.EventError](wrapped)
	require.True(t, ok)
	assert.Equal(t, callbacks.KindInsufficientPermissions, unwrapped.Kind)
	assert.Equal(t, mock.TestGuildName, unwrapped.Guild.Name)
	assert.Equal(t, mock.TestUserName, unwrapped.Member.User.Username)

	_, ok = errors.AsType[*callbacks.EventError](errors.New(wrapMsg))
	require.False(t, ok)
}
