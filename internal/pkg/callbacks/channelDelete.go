package callbacks

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

const (
	channelDelete           = "ChannelDelete"
	channelDeleteEventError = "Unable to process event: " + channelDelete
)

// ChannelDelete is the callback function for the ChannelDelete event from Discord.
func (handler *Handler) ChannelDelete(session *discordgo.Session, channel *discordgo.ChannelDelete) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), handler.ContextTimeout)
	defer cancelCtx()

	if channel.Type != discordgo.ChannelTypeGuildVoice {
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		handler.Log.WithError(err).Error(channelDeleteEventError)
		return
	}

	for _, role := range guild.Roles {
		if role.Name != handler.RolePrefix+" "+channel.Name {
			continue
		}

		err = session.GuildRoleDeleteWithContext(ctx, channel.GuildID, role.ID)
		if err != nil {
			handler.Log.WithError(err).Error(channelDeleteEventError)
			return
		}

		err = session.State.RoleRemove(channel.GuildID, role.ID)
		if err != nil {
			handler.Log.WithError(err).Error(channelDeleteEventError)
			return
		}

		handler.Log.WithField("role", role.Name).Debug("Deleted Ephemeral Role")

		return
	}
}
