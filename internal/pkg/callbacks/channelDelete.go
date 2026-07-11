package callbacks

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
)

const channelDeleteEventError = unableToProcessEvent + "ChannelDelete"

// ChannelDelete is the callback function for the ChannelDelete event from Discord.
//
// Processing is queued on the guild's sequencer, the same one VoiceStateUpdate
// uses, so a channel deletion and a concurrent voice-state update for the same
// guild never race on that guild's role state.
func (handler *Handler) ChannelDelete(event *events.GuildChannelDelete) {
	if event.Channel.Type() != discord.ChannelTypeGuildVoice {
		return
	}

	accepted := handler.sequencer.Submit(event.GuildID, func() {
		handler.handleChannelDelete(event)
	})
	if !accepted {
		// Unlike a dropped VoiceStateUpdate, a dropped ChannelDelete is never
		// retried by a later event, permanently leaking a role toward the
		// guild's 250-role cap. ChannelDelete is rare, so fall back to a
		// goroutine that waits for queue capacity: the read loop stays
		// unblocked and the work still runs serialized on the guild worker.
		handler.Log.Warn("guild queue full: queueing ChannelDelete asynchronously",
			"guildID", event.GuildID,
		)

		go handler.sequencer.SubmitWait(event.GuildID, func() {
			handler.handleChannelDelete(event)
		})
	}
}

func (handler *Handler) handleChannelDelete(event *events.GuildChannelDelete) {
	client := event.Client()
	roleName := handler.RoleNameFromChannel(event.Channel.Name())

	var (
		roleID snowflake.ID
		found  bool
	)

	// Resolve the target role before mutating the cache: the cache's Roles
	// iterator holds a lock for the duration of the range, so removing a role
	// inside the loop would deadlock.
	for role := range client.Caches.Roles(event.GuildID) {
		if role.Name == roleName {
			roleID = role.ID
			found = true

			break
		}
	}

	if !found {
		return
	}

	ctx, cancel := operations.RequestContext()
	defer cancel()

	if err := client.Rest.DeleteRole(event.GuildID, roleID, rest.WithCtx(ctx)); err != nil {
		handler.Log.Error(channelDeleteEventError, "error", err)
		return
	}

	client.Caches.RemoveRole(event.GuildID, roleID)
}
