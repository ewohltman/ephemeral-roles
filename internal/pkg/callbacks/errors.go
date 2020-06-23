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
	errMsg := memberNotFoundMessage

	if mnf.guild != nil {
		errMsg = fmt.Sprintf("%s in guild %q", errMsg, mnf.guild.Name)
	}

	if mnf.err != nil {
		errMsg = fmt.Sprintf("%s: %s", errMsg, mnf.err)
	}

	return errMsg
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
