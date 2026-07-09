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

func testChannel() discord.GuildChannel {
	return discord.GuildVoiceChannel{}
}

func TestRoleNotFound_Error(t *testing.T) {
	t.Parallel()

	rnf := &callbacks.RoleNotFoundError{}

	assert.NotEmpty(t, rnf.Error())
}

func TestMemberNotFound_Is(t *testing.T) {
	t.Parallel()

	mnf := &callbacks.MemberNotFoundError{}

	require.NotErrorIs(t, nil, &callbacks.MemberNotFoundError{})
	require.NotErrorIs(t, errors.New(wrapMsg), &callbacks.MemberNotFoundError{})
	require.ErrorIs(t, mnf, &callbacks.MemberNotFoundError{})
}

func TestMemberNotFound_Unwrap(t *testing.T) {
	t.Parallel()

	wrappedErr := errors.New(wrapMsg)

	mnf := &callbacks.MemberNotFoundError{Err: wrappedErr}

	require.ErrorIs(t, mnf.Unwrap(), wrappedErr)
}

func TestMemberNotFound_Error(t *testing.T) {
	t.Parallel()

	mnf := &callbacks.MemberNotFoundError{}
	expectedErrMsg := callbacks.MemberNotFoundMessage

	require.EqualError(t, mnf, expectedErrMsg)

	mnf.Err = errors.New(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	require.EqualError(t, mnf, expectedErrMsg)
}

func TestMemberNotFound_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discord.Guild{Name: mock.TestGuildName}
	mnf := &callbacks.MemberNotFoundError{Guild: expected}
	actual := mnf.InGuild()

	require.Equal(t, expected, actual)
}

func TestMemberNotFound_ForMember(t *testing.T) {
	t.Parallel()

	var expected *discord.Member

	mnf := &callbacks.MemberNotFoundError{}
	actual := mnf.ForMember()

	require.Equal(t, expected, actual)
}

func TestMemberNotFound_InChannel(t *testing.T) {
	t.Parallel()

	var expected discord.GuildChannel

	mnf := &callbacks.MemberNotFoundError{}
	actual := mnf.InChannel()

	require.Equal(t, expected, actual)
}

func TestChannelNotFound_Is(t *testing.T) {
	t.Parallel()

	cnf := &callbacks.ChannelNotFoundError{}

	require.NotErrorIs(t, nil, &callbacks.ChannelNotFoundError{})
	require.NotErrorIs(t, errors.New(wrapMsg), &callbacks.ChannelNotFoundError{})
	require.ErrorIs(t, cnf, &callbacks.ChannelNotFoundError{})
}

func TestChannelNotFound_Unwrap(t *testing.T) {
	t.Parallel()

	wrappedErr := errors.New(wrapMsg)

	cnf := &callbacks.ChannelNotFoundError{Err: wrappedErr}

	require.ErrorIs(t, cnf.Unwrap(), wrappedErr)
}

func TestChannelNotFound_Error(t *testing.T) {
	t.Parallel()

	cnf := &callbacks.ChannelNotFoundError{}
	expectedErrMsg := callbacks.ChannelNotFoundMessage

	require.EqualError(t, cnf, expectedErrMsg)

	cnf.Err = errors.New(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	require.EqualError(t, cnf, expectedErrMsg)
}

func TestChannelNotFound_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discord.Guild{Name: mock.TestGuildName}
	cnf := &callbacks.ChannelNotFoundError{Guild: expected}
	actual := cnf.InGuild()

	require.Equal(t, expected, actual)
}

func TestChannelNotFound_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discord.Member{User: discord.User{Username: mock.TestUserName}}
	cnf := &callbacks.ChannelNotFoundError{Member: expected}
	actual := cnf.ForMember()

	require.Equal(t, expected, actual)
}

func TestChannelNotFound_InChannel(t *testing.T) {
	t.Parallel()

	var expected discord.GuildChannel

	cnf := &callbacks.ChannelNotFoundError{}
	actual := cnf.InChannel()

	require.Equal(t, expected, actual)
}

func TestInsufficientPermission_Is(t *testing.T) {
	t.Parallel()

	inp := &callbacks.InsufficientPermissionsError{}

	require.NotErrorIs(t, nil, &callbacks.InsufficientPermissionsError{})
	require.NotErrorIs(t, errors.New(wrapMsg), &callbacks.InsufficientPermissionsError{})
	require.ErrorIs(t, inp, &callbacks.InsufficientPermissionsError{})
}

func TestInsufficientPermission_Unwrap(t *testing.T) {
	t.Parallel()

	inp := &callbacks.InsufficientPermissionsError{}

	require.NoError(t, inp.Unwrap())
}

func TestInsufficientPermission_Error(t *testing.T) {
	t.Parallel()

	inp := &callbacks.InsufficientPermissionsError{}
	expectedErrMsg := callbacks.InsufficientPermissionMessage

	require.EqualError(t, inp, expectedErrMsg)

	inp.Err = errors.New(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	require.EqualError(t, inp, expectedErrMsg)
}

func TestInsufficientPermissions_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discord.Guild{Name: mock.TestGuildName}
	inp := &callbacks.InsufficientPermissionsError{Guild: expected}
	actual := inp.InGuild()

	require.Equal(t, expected, actual)
}

func TestInsufficientPermissions_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discord.Member{User: discord.User{Username: mock.TestUserName}}
	inp := &callbacks.InsufficientPermissionsError{Member: expected}
	actual := inp.ForMember()

	require.Equal(t, expected, actual)
}

func TestInsufficientPermissions_InChannel(t *testing.T) {
	t.Parallel()

	expected := testChannel()
	inp := &callbacks.InsufficientPermissionsError{Channel: expected}
	actual := inp.InChannel()

	require.Equal(t, expected, actual)
}

func TestMaxNumberOfRoles_Is(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.MaxNumberOfRolesError{}

	require.NotErrorIs(t, nil, &callbacks.MaxNumberOfRolesError{})
	require.NotErrorIs(t, errors.New(wrapMsg), &callbacks.MaxNumberOfRolesError{})
	require.ErrorIs(t, mnr, &callbacks.MaxNumberOfRolesError{})
}

func TestMaxNumberOfRoles_Unwrap(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.MaxNumberOfRolesError{}

	require.NoError(t, mnr.Unwrap())
}

func TestMaxNumberOfRoles_Error(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.MaxNumberOfRolesError{}
	expectedErrMsg := callbacks.MaxNumberOfRolesMessage

	require.EqualError(t, mnr, expectedErrMsg)

	mnr.Err = errors.New(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	require.EqualError(t, mnr, expectedErrMsg)
}

func TestMaxNumberOfRoles_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discord.Guild{Name: mock.TestGuildName}
	mnr := &callbacks.MaxNumberOfRolesError{Guild: expected}
	actual := mnr.InGuild()

	require.Equal(t, expected, actual)
}

func TestMaxNumberOfRoles_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discord.Member{User: discord.User{Username: mock.TestUserName}}
	mnr := &callbacks.MaxNumberOfRolesError{Member: expected}
	actual := mnr.ForMember()

	require.Equal(t, expected, actual)
}

func TestMaxNumberOfRoles_InChannel(t *testing.T) {
	t.Parallel()

	expected := testChannel()
	mnr := &callbacks.MaxNumberOfRolesError{Channel: expected}
	actual := mnr.InChannel()

	require.Equal(t, expected, actual)
}

func TestDeadlineExceeded_Is(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.DeadlineExceededError{}

	require.NotErrorIs(t, nil, &callbacks.DeadlineExceededError{})
	require.NotErrorIs(t, errors.New(wrapMsg), &callbacks.DeadlineExceededError{})
	require.ErrorIs(t, mnr, &callbacks.DeadlineExceededError{})
}

func TestDeadlineExceeded_Unwrap(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.DeadlineExceededError{}

	require.NoError(t, mnr.Unwrap())
}

func TestDeadlineExceeded_Error(t *testing.T) {
	t.Parallel()

	mnr := &callbacks.DeadlineExceededError{}
	expectedErrMsg := callbacks.DeadlineExceededMessage

	require.EqualError(t, mnr, expectedErrMsg)

	mnr.Err = errors.New(wrapMsg)
	expectedErrMsg = fmt.Sprintf("%s: %s", expectedErrMsg, wrapMsg)

	require.EqualError(t, mnr, expectedErrMsg)
}

func TestDeadlineExceeded_InGuild(t *testing.T) {
	t.Parallel()

	expected := &discord.Guild{Name: mock.TestGuildName}
	mnr := &callbacks.DeadlineExceededError{Guild: expected}
	actual := mnr.InGuild()

	require.Equal(t, expected, actual)
}

func TestDeadlineExceeded_ForMember(t *testing.T) {
	t.Parallel()

	expected := &discord.Member{User: discord.User{Username: mock.TestUserName}}
	mnr := &callbacks.DeadlineExceededError{Member: expected}
	actual := mnr.ForMember()

	require.Equal(t, expected, actual)
}

func TestDeadlineExceeded_InChannel(t *testing.T) {
	t.Parallel()

	expected := testChannel()
	mnr := &callbacks.DeadlineExceededError{Channel: expected}
	actual := mnr.InChannel()

	require.Equal(t, expected, actual)
}
