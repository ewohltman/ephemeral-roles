package callbacks

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

const readyEventError = unableToProcessEvent + "Ready"

// Ready is the callback function for the Ready event from Discord.
func (handler *Handler) Ready(event *events.Ready) {
	handler.ReadyCounter.Inc()

	err := event.Client().SetPresenceForShard(context.Background(), event.ShardID(),
		gateway.WithOnlineStatus(discord.OnlineStatusOnline),
		gateway.WithWatchingActivity("voice channels"),
	)
	if err != nil {
		handler.Log.Error(readyEventError, "error", err)
	}
}
