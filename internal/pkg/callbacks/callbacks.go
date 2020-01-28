// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

const userNotFoundError = "user not found in guild members"

// Config contains fields for the callback methods.
type Config struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}

type userNotFound struct{}

func (unf *userNotFound) Error() string {
	return userNotFoundError
}
