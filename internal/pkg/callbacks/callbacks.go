// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"fmt"
	"log/slog"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const unableToProcessEvent = "unable to process event: "

// OperationsGateway is an interface abstraction for processing operations
// requests.
type OperationsGateway interface {
	Process(request *operations.Request) <-chan singleflight.Result
}

// Handler contains fields for the callback methods attached to it.
type Handler struct {
	Log                     *slog.Logger
	RolePrefix              string
	RoleColor               int
	JaegerTracer            opentracing.Tracer
	ReadyCounter            prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
	OperationsGateway       OperationsGateway
}

// RoleNameFromChannel returns the name of a role for a channel, with the bot
// keyword prefixed.
func (handler *Handler) RoleNameFromChannel(channelName string) string {
	return fmt.Sprintf("%s %s", handler.RolePrefix, channelName)
}
