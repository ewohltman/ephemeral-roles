package callbacks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// GuildMemberAdd is the callback function for the GuildMemberAdd event from Discord
func GuildMemberAdd(s *discordgo.Session, memberAddEvent *discordgo.GuildMemberAdd) {
	if memberAddEvent.User.Bot && memberAddEvent.User.ID == s.State.User.ID {
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
}
