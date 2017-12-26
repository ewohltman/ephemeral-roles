package callbacks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Ready is the callback function for the "ready" event from Discord
func Ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithFields(logrus.Fields{
		"guilds": len(event.Guilds),
	}).Infof(BOT_NAME + " started up")

	s.UpdateStatus(0, BOT_KEYWORD) // Set the Discord "playing" status
}
