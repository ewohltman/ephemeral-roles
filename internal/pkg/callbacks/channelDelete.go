package callbacks

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

const channelDeleteEventError = unableToProcessEvent + "ChannelDelete"

// ChannelDelete is the callback function for the ChannelDelete event from Discord.
func (handler *Handler) ChannelDelete(session *discordgo.Session, channel *discordgo.ChannelDelete) {
	if channel.Type != discordgo.ChannelTypeGuildVoice {
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		handler.Log.Error(channelDeleteEventError, "error", err)
		return
	}

	for _, role := range guild.Roles {
		if role.Name != handler.RoleNameFromChannel(channel.Name) {
			continue
		}

		if err := session.GuildRoleDelete(channel.GuildID, role.ID); err != nil {
			handler.Log.Error(channelDeleteEventError, "error", err)
			return
		}

		if err := session.State.RoleRemove(channel.GuildID, role.ID); err != nil && !errors.Is(err, discordgo.ErrStateNotFound) {
			handler.Log.Error(channelDeleteEventError, "error", err)
			return
		}

		return
	}
}
