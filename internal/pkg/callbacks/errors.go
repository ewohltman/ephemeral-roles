package callbacks

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

// Error messages.
const (
	MemberNotFoundMessage         = "member not found"
	ChannelNotFoundMessage        = "channel not found"
	InsufficientPermissionMessage = "insufficient permissions"
	MaxNumberOfRolesMessage       = "max number of roles"
	DeadlineExceededMessage       = "deadline exceeded"
)

// ErrorKind classifies the failure modes encountered when processing a
// callback event.
type ErrorKind int

// ErrorKind enumerations.
const (
	KindMemberNotFound ErrorKind = iota
	KindChannelNotFound
	KindInsufficientPermissions
	KindMaxNumberOfRoles
	KindDeadlineExceeded
)

// Message returns the error message for the ErrorKind.
func (kind ErrorKind) Message() string {
	switch kind {
	case KindMemberNotFound:
		return MemberNotFoundMessage
	case KindChannelNotFound:
		return ChannelNotFoundMessage
	case KindInsufficientPermissions:
		return InsufficientPermissionMessage
	case KindMaxNumberOfRoles:
		return MaxNumberOfRolesMessage
	case KindDeadlineExceeded:
		return DeadlineExceededMessage
	default:
		return "unknown error"
	}
}

// EventError is a typed error for failures processing callback events. It
// carries the guild, member, and channel context available at the point of
// failure so handlers can branch on Kind and attach structured log fields.
type EventError struct {
	Kind    ErrorKind
	Guild   *discord.Guild
	Member  *discord.Member
	Channel discord.GuildChannel
	Err     error
}

// Error satisfies the error interface for EventError.
func (eventErr *EventError) Error() string {
	if eventErr.Err != nil {
		return fmt.Sprintf("%s: %s", eventErr.Kind.Message(), eventErr.Err)
	}

	return eventErr.Kind.Message()
}

// Unwrap returns an error wrapped by EventError.
func (eventErr *EventError) Unwrap() error {
	return eventErr.Err
}
