package callbacks

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const channelDeleteEventError = unableToProcessEvent + "ChannelDelete"

// ChannelDelete is the callback function for the ChannelDelete event from Discord.
func (handler *Handler) ChannelDelete(event *events.GuildChannelDelete) {
	if event.Channel.Type() != discord.ChannelTypeGuildVoice {
		return
	}

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

	if err := client.Rest.DeleteRole(event.GuildID, roleID); err != nil {
		handler.Log.Error(channelDeleteEventError, "error", err)
		return
	}

	client.Caches.RemoveRole(event.GuildID, roleID)
}
