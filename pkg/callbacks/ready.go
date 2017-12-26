package callbacks

import (
	"github.com/ewohltman/discordgo"
	"github.com/sirupsen/logrus"
)

// Ready is the callback function for the Ready event from Discord
func Ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithFields(logrus.Fields{
		"guildCount": len(event.Guilds),
	}).Infof(BOT_NAME + " Ready event")

	idleSince := 0

	usd := discordgo.UpdateStatusData{
		IdleSince: &idleSince,
		Game: &discordgo.Game{
			Name: BOT_KEYWORD,
			Type: discordgo.GameTypeWatching,
		},
		AFK:    false,
		Status: "online",
	}

	err := s.UpdateStatusComplex(usd)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"UpdateStatusData": usd,
		}).Errorf("Error updating complex status")
	}
}
