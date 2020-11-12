// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

// Handler contains fields for the callback methods attached to it.
type Handler struct {
	Log                     logging.Interface
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	RoleColor               int
	JaegerTracer            opentracing.Tracer
	ContextTimeout          time.Duration
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}
