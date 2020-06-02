package callbacks

import "github.com/bwmarrin/discordgo"

const (
	channelDelete = "ChannelDelete"

	channelDeleteEventError = "Unable to process event: " + channelDelete
)

// ChannelDelete is the callback function for the ChannelDelete event from Discord.
func (config *Config) ChannelDelete(session *discordgo.Session, channel *discordgo.ChannelDelete) {
	if channel.Type != discordgo.ChannelTypeGuildVoice {
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		config.Log.WithError(err).Error(channelDeleteEventError)
		return
	}

	for _, role := range guild.Roles {
		if role.Name != config.RolePrefix+" "+channel.Name {
			continue
		}

		err = session.State.RoleRemove(channel.GuildID, role.ID)
		if err != nil {
			config.Log.WithError(err).Error(channelDeleteEventError)
			return
		}

		err = session.GuildRoleDelete(channel.GuildID, role.ID)
		if err != nil {
			config.Log.WithError(err).Error(channelDeleteEventError)

			err = session.State.RoleAdd(channel.GuildID, role)
			if err != nil {
				config.Log.WithError(err).Error(channelDeleteEventError)
			}

			return
		}

		config.Log.WithField("role", role.Name).Debug("Deleted Ephemeral Role")

		return
	}
}
