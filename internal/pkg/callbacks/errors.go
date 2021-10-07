package callbacks

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Error messages.
const (
	MemberNotFoundMessage         = "member not found"
	ChannelNotFoundMessage        = "channel not found"
	RoleNotFoundMessage           = "role not found"
	InsufficientPermissionMessage = "insufficient permissions"
	MaxNumberOfRolesMessage       = "max number of roles"
	DeadlineExceededMessage       = "deadline exceeded"
)

// CallbackError embeds the error interface with additional methods to provide
// metadata for the error.
type CallbackError interface {
	error
	Is(error) bool
	Unwrap() error
	InGuild() *discordgo.Guild
	ForMember() *discordgo.Member
	InChannel() *discordgo.Channel
}

// MemberNotFound represents an error when a member is not found.
type MemberNotFound struct {
	Guild *discordgo.Guild
	Err   error
}

// Is allows MemberNotFound to be compared with errors.Is.
func (_ *MemberNotFound) Is(target error) bool {
	_, ok := target.(*MemberNotFound)
	return ok
}

// Unwrap returns an error wrapped by MemberNotFound.
func (mnf *MemberNotFound) Unwrap() error {
	return mnf.Err
}

// Error satisfies the errors interface for MemberNotFound.
func (mnf *MemberNotFound) Error() string {
	if mnf.Err != nil {
		return fmt.Sprintf("%s: %s", MemberNotFoundMessage, mnf.Err)
	}

	return MemberNotFoundMessage
}

// InGuild returns guild metadata for MemberNotFound.
func (mnf *MemberNotFound) InGuild() *discordgo.Guild {
	return mnf.Guild
}

// ForMember satisfies the CallbackError interface for MemberNotFound.
func (_ *MemberNotFound) ForMember() *discordgo.Member {
	return nil
}

// InChannel satisfies the CallbackError interface for MemberNotFound.
func (_ *MemberNotFound) InChannel() *discordgo.Channel {
	return nil
}

// ChannelNotFound represents an error when a channel is not found.
type ChannelNotFound struct {
	Guild  *discordgo.Guild
	Member *discordgo.Member
	Err    error
}

// Is allows ChannelNotFound to be compared with errors.Is.
func (_ *ChannelNotFound) Is(target error) bool {
	_, ok := target.(*ChannelNotFound)
	return ok
}

// Unwrap returns an error wrapped by ChannelNotFound.
func (cnf *ChannelNotFound) Unwrap() error {
	return cnf.Err
}

// Error satisfies the errors interface for ChannelNotFound.
func (cnf *ChannelNotFound) Error() string {
	if cnf.Err != nil {
		return fmt.Sprintf("%s: %s", ChannelNotFoundMessage, cnf.Err)
	}

	return ChannelNotFoundMessage
}

// InGuild returns guild metadata for ChannelNotFound.
func (cnf *ChannelNotFound) InGuild() *discordgo.Guild {
	return cnf.Guild
}

// ForMember returns member metadata for ChannelNotFound.
func (cnf *ChannelNotFound) ForMember() *discordgo.Member {
	return cnf.Member
}

// InChannel satisfies the CallbackError interface for ChannelNotFound.
func (_ *ChannelNotFound) InChannel() *discordgo.Channel {
	return nil
}

// RoleNotFound represents an error for when the bot fails to find a role.
type RoleNotFound struct{}

// Error satisfies the errors interface for RoleNotFound.
func (_ *RoleNotFound) Error() string {
	return RoleNotFoundMessage
}

// InsufficientPermissions represents an error for when the bot lacks role
// privileges to perform an operation.
type InsufficientPermissions struct {
	Guild   *discordgo.Guild
	Member  *discordgo.Member
	Channel *discordgo.Channel
	Err     error
}

// Is allows InsufficientPermissions to be compared with errors.Is.
func (_ *InsufficientPermissions) Is(target error) bool {
	_, ok := target.(*InsufficientPermissions)
	return ok
}

// Unwrap returns an error wrapped by InsufficientPermissions.
func (inp *InsufficientPermissions) Unwrap() error {
	return inp.Err
}

// Error satisfies the errors interface for InsufficientPermissions.
func (inp *InsufficientPermissions) Error() string {
	if inp.Err != nil {
		return fmt.Sprintf("%s: %s", InsufficientPermissionMessage, inp.Err)
	}

	return InsufficientPermissionMessage
}

// InGuild returns guild metadata for InsufficientPermissions.
func (inp *InsufficientPermissions) InGuild() *discordgo.Guild {
	return inp.Guild
}

// ForMember returns member metadata for InsufficientPermissions.
func (inp *InsufficientPermissions) ForMember() *discordgo.Member {
	return inp.Member
}

// InChannel returns channel metadata for InsufficientPermissions.
func (inp *InsufficientPermissions) InChannel() *discordgo.Channel {
	return inp.Channel
}

// MaxNumberOfRoles represents an error for when a guild already has the max
// number of roles allowed.
type MaxNumberOfRoles struct {
	Guild   *discordgo.Guild
	Member  *discordgo.Member
	Channel *discordgo.Channel
	Err     error
}

// Is allows MaxNumberOfRoles to be compared with errors.Is.
func (_ *MaxNumberOfRoles) Is(target error) bool {
	_, ok := target.(*MaxNumberOfRoles)
	return ok
}

// Unwrap returns an error wrapped by MaxNumberOfRoles.
func (mnr *MaxNumberOfRoles) Unwrap() error {
	return mnr.Err
}

// Error satisfies the errors interface for MaxNumberOfRoles.
func (mnr *MaxNumberOfRoles) Error() string {
	if mnr.Err != nil {
		return fmt.Sprintf("%s: %s", MaxNumberOfRolesMessage, mnr.Err)
	}

	return MaxNumberOfRolesMessage
}

// InGuild returns guild metadata for MaxNumberOfRoles.
func (mnr *MaxNumberOfRoles) InGuild() *discordgo.Guild {
	return mnr.Guild
}

// ForMember returns member metadata for MaxNumberOfRoles.
func (mnr *MaxNumberOfRoles) ForMember() *discordgo.Member {
	return mnr.Member
}

// InChannel returns channel metadata for MaxNumberOfRoles.
func (mnr *MaxNumberOfRoles) InChannel() *discordgo.Channel {
	return mnr.Channel
}

// DeadlineExceeded represents an error for when a context deadline has been
// exceeded.
type DeadlineExceeded struct {
	Guild   *discordgo.Guild
	Member  *discordgo.Member
	Channel *discordgo.Channel
	Err     error
}

// Error satisfies the errors interface for DeadlineExceeded.
func (de *DeadlineExceeded) Error() string {
	if de.Err != nil {
		return fmt.Sprintf("%s: %s", DeadlineExceededMessage, de.Err)
	}

	return DeadlineExceededMessage
}

// Is allows DeadlineExceeded to be compared with errors.Is.
func (_ *DeadlineExceeded) Is(target error) bool {
	_, ok := target.(*DeadlineExceeded)
	return ok
}

// Unwrap returns an error wrapped by DeadlineExceeded.
func (de *DeadlineExceeded) Unwrap() error {
	return de.Err
}

// InGuild returns guild metadata for DeadlineExceeded.
func (de *DeadlineExceeded) InGuild() *discordgo.Guild {
	return de.Guild
}

// ForMember returns member metadata for DeadlineExceeded.
func (de *DeadlineExceeded) ForMember() *discordgo.Member {
	return de.Member
}

// InChannel returns channel metadata for DeadlineExceeded.
func (de *DeadlineExceeded) InChannel() *discordgo.Channel {
	return de.Channel
}
