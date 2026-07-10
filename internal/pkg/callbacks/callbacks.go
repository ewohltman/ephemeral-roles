// Package callbacks provides callback implementations for Discord API events.
package callbacks

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/prometheus/client_golang/prometheus"
)

const unableToProcessEvent = "unable to process event: "

// OperationsGateway is an interface abstraction for processing operations
// requests.
type OperationsGateway interface {
	CreateRole(guildID snowflake.ID, roleName string, roleColor int) (discord.Role, error)
}

// Handler contains fields for the callback methods attached to it.
type Handler struct {
	Log                     *slog.Logger
	RolePrefix              string
	RoleColor               int
	ReadyCounter            prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
	OperationsGateway       OperationsGateway

	sequencer guildSequencer
}

// Flush blocks until any Discord role work already queued for guildID (from
// VoiceStateUpdate or ChannelDelete) has completed.
func (handler *Handler) Flush(guildID snowflake.ID) {
	handler.sequencer.Flush(guildID)
}

// RoleNameFromChannel returns the name of a role for a channel, with the bot
// keyword prefixed.
func (handler *Handler) RoleNameFromChannel(channelName string) string {
	return handler.RolePrefix + " " + channelName
}
