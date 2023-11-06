package callbacks

import (
	"errors"
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
	Is(err error) bool
	Unwrap() error
	InGuild() *discordgo.Guild
	ForMember() *discordgo.Member
	InChannel() *discordgo.Channel
}

// MemberNotFoundError represents an error when a member is not found.
type MemberNotFoundError struct {
	Guild *discordgo.Guild
	Err   error
}

// Is allows MemberNotFoundError to be compared with errors.Is.
func (mnf *MemberNotFoundError) Is(target error) bool {
	return errors.As(target, &mnf)
}

// Unwrap returns an error wrapped by MemberNotFoundError.
func (mnf *MemberNotFoundError) Unwrap() error {
	return mnf.Err
}

// Error satisfies the errors interface for MemberNotFoundError.
func (mnf *MemberNotFoundError) Error() string {
	if mnf.Err != nil {
		return fmt.Sprintf("%s: %s", MemberNotFoundMessage, mnf.Err)
	}

	return MemberNotFoundMessage
}

// InGuild returns guild metadata for MemberNotFoundError.
func (mnf *MemberNotFoundError) InGuild() *discordgo.Guild {
	return mnf.Guild
}

// ForMember satisfies the CallbackError interface for MemberNotFoundError.
func (*MemberNotFoundError) ForMember() *discordgo.Member {
	return nil
}

// InChannel satisfies the CallbackError interface for MemberNotFoundError.
func (*MemberNotFoundError) InChannel() *discordgo.Channel {
	return nil
}

// ChannelNotFoundError represents an error when a channel is not found.
type ChannelNotFoundError struct {
	Guild  *discordgo.Guild
	Member *discordgo.Member
	Err    error
}

// Is allows ChannelNotFoundError to be compared with errors.Is.
func (cnf *ChannelNotFoundError) Is(target error) bool {
	return errors.As(target, &cnf)
}

// Unwrap returns an error wrapped by ChannelNotFoundError.
func (cnf *ChannelNotFoundError) Unwrap() error {
	return cnf.Err
}

// Error satisfies the errors interface for ChannelNotFoundError.
func (cnf *ChannelNotFoundError) Error() string {
	if cnf.Err != nil {
		return fmt.Sprintf("%s: %s", ChannelNotFoundMessage, cnf.Err)
	}

	return ChannelNotFoundMessage
}

// InGuild returns guild metadata for ChannelNotFoundError.
func (cnf *ChannelNotFoundError) InGuild() *discordgo.Guild {
	return cnf.Guild
}

// ForMember returns member metadata for ChannelNotFoundError.
func (cnf *ChannelNotFoundError) ForMember() *discordgo.Member {
	return cnf.Member
}

// InChannel satisfies the CallbackError interface for ChannelNotFoundError.
func (*ChannelNotFoundError) InChannel() *discordgo.Channel {
	return nil
}

// RoleNotFoundError represents an error for when the bot fails to find a role.
type RoleNotFoundError struct{}

// Error satisfies the errors interface for RoleNotFoundError.
func (*RoleNotFoundError) Error() string {
	return RoleNotFoundMessage
}

// InsufficientPermissionsError represents an error for when the bot lacks role
// privileges to perform an operation.
type InsufficientPermissionsError struct {
	Guild   *discordgo.Guild
	Member  *discordgo.Member
	Channel *discordgo.Channel
	Err     error
}

// Is allows InsufficientPermissionsError to be compared with errors.Is.
func (inp *InsufficientPermissionsError) Is(target error) bool {
	return errors.As(target, &inp)
}

// Unwrap returns an error wrapped by InsufficientPermissionsError.
func (inp *InsufficientPermissionsError) Unwrap() error {
	return inp.Err
}

// Error satisfies the errors interface for InsufficientPermissionsError.
func (inp *InsufficientPermissionsError) Error() string {
	if inp.Err != nil {
		return fmt.Sprintf("%s: %s", InsufficientPermissionMessage, inp.Err)
	}

	return InsufficientPermissionMessage
}

// InGuild returns guild metadata for InsufficientPermissionsError.
func (inp *InsufficientPermissionsError) InGuild() *discordgo.Guild {
	return inp.Guild
}

// ForMember returns member metadata for InsufficientPermissionsError.
func (inp *InsufficientPermissionsError) ForMember() *discordgo.Member {
	return inp.Member
}

// InChannel returns channel metadata for InsufficientPermissionsError.
func (inp *InsufficientPermissionsError) InChannel() *discordgo.Channel {
	return inp.Channel
}

// MaxNumberOfRolesError represents an error for when a guild already has the max
// number of roles allowed.
type MaxNumberOfRolesError struct {
	Guild   *discordgo.Guild
	Member  *discordgo.Member
	Channel *discordgo.Channel
	Err     error
}

// Is allows MaxNumberOfRolesError to be compared with errors.Is.
func (mnr *MaxNumberOfRolesError) Is(target error) bool {
	return errors.As(target, &mnr)
}

// Unwrap returns an error wrapped by MaxNumberOfRolesError.
func (mnr *MaxNumberOfRolesError) Unwrap() error {
	return mnr.Err
}

// Error satisfies the errors interface for MaxNumberOfRolesError.
func (mnr *MaxNumberOfRolesError) Error() string {
	if mnr.Err != nil {
		return fmt.Sprintf("%s: %s", MaxNumberOfRolesMessage, mnr.Err)
	}

	return MaxNumberOfRolesMessage
}

// InGuild returns guild metadata for MaxNumberOfRolesError.
func (mnr *MaxNumberOfRolesError) InGuild() *discordgo.Guild {
	return mnr.Guild
}

// ForMember returns member metadata for MaxNumberOfRolesError.
func (mnr *MaxNumberOfRolesError) ForMember() *discordgo.Member {
	return mnr.Member
}

// InChannel returns channel metadata for MaxNumberOfRolesError.
func (mnr *MaxNumberOfRolesError) InChannel() *discordgo.Channel {
	return mnr.Channel
}

// DeadlineExceededError represents an error for when a context deadline has been
// exceeded.
type DeadlineExceededError struct {
	Guild   *discordgo.Guild
	Member  *discordgo.Member
	Channel *discordgo.Channel
	Err     error
}

// Error satisfies the errors interface for DeadlineExceededError.
func (de *DeadlineExceededError) Error() string {
	if de.Err != nil {
		return fmt.Sprintf("%s: %s", DeadlineExceededMessage, de.Err)
	}

	return DeadlineExceededMessage
}

// Is allows DeadlineExceededError to be compared with errors.Is.
func (de *DeadlineExceededError) Is(target error) bool {
	return errors.As(target, &de)
}

// Unwrap returns an error wrapped by DeadlineExceededError.
func (de *DeadlineExceededError) Unwrap() error {
	return de.Err
}

// InGuild returns guild metadata for DeadlineExceededError.
func (de *DeadlineExceededError) InGuild() *discordgo.Guild {
	return de.Guild
}

// ForMember returns member metadata for DeadlineExceededError.
func (de *DeadlineExceededError) ForMember() *discordgo.Member {
	return de.Member
}

// InChannel returns channel metadata for DeadlineExceededError.
func (de *DeadlineExceededError) InChannel() *discordgo.Channel {
	return de.Channel
}
