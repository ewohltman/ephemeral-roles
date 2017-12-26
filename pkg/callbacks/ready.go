package callbacks

import (
	"github.com/ewohltman/discordgo"
	"github.com/sirupsen/logrus"
)

var guildCount int

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

		// Fall-back set the Discord "playing" status
		s.UpdateStatus(0, BOT_KEYWORD)
	}
}

func newGuildMonitor(s *discordgo.Session) {
	for true {
		// Guild added
		if len(s.State.Guilds) > guildCount {
			log.WithFields(logrus.Fields{
				"guildCount": guildCount,
			}).Infof(BOT_NAME + " Ready event")

			continue
		}
	}
}
