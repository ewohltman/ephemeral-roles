package callbacks

import "fmt"

const (
	memberNotFoundMessage  = "guild member not found"
	channelNotFoundMessage = "channel not found"
)

type memberNotFound struct {
	err error
}

func (mnf *memberNotFound) Is(target error) bool {
	_, ok := target.(*memberNotFound)
	return ok
}

func (mnf *memberNotFound) UnWrap() error {
	return mnf.err
}

func (mnf *memberNotFound) Error() string {
	if mnf.err != nil {
		return fmt.Sprintf("%s: %s", memberNotFoundMessage, mnf.err)
	}

	return memberNotFoundMessage
}

type channelNotFound struct {
	err error
}

func (cnf *channelNotFound) Is(target error) bool {
	_, ok := target.(*channelNotFound)
	return ok
}

func (cnf *channelNotFound) UnWrap() error {
	return cnf.err
}

func (cnf *channelNotFound) Error() string {
	if cnf.err != nil {
		return fmt.Sprintf("%s: %s", channelNotFoundMessage, cnf.err)
	}

	return channelNotFoundMessage
}
