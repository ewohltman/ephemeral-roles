package callbacks

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

const (
	channelDelete           = "ChannelDelete"
	channelDeleteEventError = "Unable to process event: " + channelDelete
)

// ChannelDelete is the callback function for the ChannelDelete event from Discord.
func (handler *Handler) ChannelDelete(session *discordgo.Session, channel *discordgo.ChannelDelete) {
	if channel.Type != discordgo.ChannelTypeGuildVoice {
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		handler.Log.WithError(err).Error(channelDeleteEventError)

		return
	}

	for _, role := range guild.Roles {
		if role.Name != handler.RoleNameFromChannel(channel.Name) {
			continue
		}

		err = session.GuildRoleDelete(channel.GuildID, role.ID)
		if err != nil {
			handler.Log.WithError(err).Error(channelDeleteEventError)

			return
		}

		err = session.State.RoleRemove(channel.GuildID, role.ID)
		if err != nil && !errors.Is(err, discordgo.ErrStateNotFound) {
			handler.Log.WithError(err).Error(channelDeleteEventError)

			return
		}

		return
	}
}
