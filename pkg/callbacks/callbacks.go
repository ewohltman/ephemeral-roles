package callbacks

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

type Config struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}
