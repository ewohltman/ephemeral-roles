// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

const userNotFoundErrorMessage = "user not found"

// Config contains fields for the callback methods.
type Config struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	RoleColor               int
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}

type userNotFoundError struct {
	err error
}

func (unf *userNotFoundError) UnWrap() error {
	return unf.err
}

func (unf *userNotFoundError) Error() string {
	if unf.err != nil {
		return fmt.Sprintf("%s: %s", userNotFoundErrorMessage, unf.err)
	}

	return userNotFoundErrorMessage
}
