package callbacks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	memberNotFoundMessage         = "member not found"
	channelNotFoundMessage        = "channel not found"
	insufficientPermissionMessage = "insufficient channel permission"
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
	var errMsg string

	if mnf.guild != nil {
		errMsg = mnf.guild.Name
	}

	errMsg = strings.TrimSpace(
		fmt.Sprintf("%s %s", errMsg, memberNotFoundMessage),
	)

	if mnf.err != nil {
		return fmt.Sprintf("%s: %s", errMsg, mnf.err)
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
	var errMsg string

	if cnf.guild != nil {
		errMsg = cnf.guild.Name
	}

	errMsg = strings.TrimSpace(
		fmt.Sprintf("%s %s", errMsg, channelNotFoundMessage),
	)

	if cnf.err != nil {
		return fmt.Sprintf("%s: %s", errMsg, cnf.err)
	}

	return errMsg
}

type insufficientPermission struct {
	guild   *discordgo.Guild
	member  *discordgo.Member
	channel *discordgo.Channel
	err     error
}

func (inp *insufficientPermission) Is(target error) bool {
	_, ok := target.(*insufficientPermission)
	return ok
}

func (inp *insufficientPermission) Unwrap() error {
	return inp.err
}

func (inp *insufficientPermission) Error() string {
	if inp.err != nil {
		return fmt.Sprintf("%s: %s", insufficientPermissionMessage, inp.err)
	}

	return insufficientPermissionMessage
}
