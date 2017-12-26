package callbacks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// BotJoin is the callback function for the "ready" event from Discord
func BotJoin(s *discordgo.Session, memberAddEvent *discordgo.GuildMemberAdd) {
	// Not a bot
	if !memberAddEvent.User.Bot {
		return
	}

	// Not our bot
	if memberAddEvent.User.ID != s.State.User.ID {
		return
	}

	guild, err := s.Guild(memberAddEvent.GuildID)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"guildID": memberAddEvent.GuildID,
		}).Errorf("Error looking up Guild ID")

		return
	}

	log.WithFields(logrus.Fields{
		"guild": guild.Name,
	}).Infof(BOT_NAME + " joined new server")
}
