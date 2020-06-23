package callbacks

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

const (
	memberNotFoundMessage         = "member not found"
	channelNotFoundMessage        = "channel not found"
	insufficientPermissionMessage = "insufficient permissions"
)

type customError interface {
	error
	Guild() *discordgo.Guild
	Member() *discordgo.Member
	Channel() *discordgo.Channel
}

type memberNotFound struct {
	guild *discordgo.Guild
	err   error
}

func (mnf *memberNotFound) Is(target error) bool {
	_, ok := target.(*memberNotFound)
	return ok
}

func (mnf *memberNotFound) Unwrap() error {
	return mnf.err
}

func (mnf *memberNotFound) Error() string {
	if mnf.err != nil {
		return fmt.Sprintf("%s: %s", memberNotFoundMessage, mnf.err)
	}

	return memberNotFoundMessage
}

func (mnf *memberNotFound) Guild() *discordgo.Guild {
	return mnf.guild
}

func (mnf *memberNotFound) Member() *discordgo.Member {
	return nil
}

func (mnf *memberNotFound) Channel() *discordgo.Channel {
	return nil
}

type channelNotFound struct {
	guild  *discordgo.Guild
	member *discordgo.Member
	err    error
}

func (cnf *channelNotFound) Is(target error) bool {
	_, ok := target.(*channelNotFound)
	return ok
}

func (cnf *channelNotFound) Unwrap() error {
	return cnf.err
}

func (cnf *channelNotFound) Error() string {
	if cnf.err != nil {
		return fmt.Sprintf("%s: %s", channelNotFoundMessage, cnf.err)
	}

	return channelNotFoundMessage
}

func (cnf *channelNotFound) Guild() *discordgo.Guild {
	return cnf.guild
}

func (cnf *channelNotFound) Member() *discordgo.Member {
	return cnf.member
}

func (cnf *channelNotFound) Channel() *discordgo.Channel {
	return nil
}

type insufficientPermissions struct {
	guild   *discordgo.Guild
	member  *discordgo.Member
	channel *discordgo.Channel
	err     error
}

func (inp *insufficientPermissions) Is(target error) bool {
	_, ok := target.(*insufficientPermissions)
	return ok
}

func (inp *insufficientPermissions) Unwrap() error {
	return inp.err
}

func (inp *insufficientPermissions) Error() string {
	if inp.err != nil {
		return fmt.Sprintf("%s: %s", insufficientPermissionMessage, inp.err)
	}

	return insufficientPermissionMessage
}

func (inp *insufficientPermissions) Guild() *discordgo.Guild {
	return inp.guild
}

func (inp *insufficientPermissions) Member() *discordgo.Member {
	return inp.member
}

func (inp *insufficientPermissions) Channel() *discordgo.Channel {
	return inp.channel
}
