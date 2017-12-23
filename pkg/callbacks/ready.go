package callbacks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Ready is the callback function for the "ready" event from Discord
func Ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithFields(logrus.Fields{
		"servers": len(event.Guilds),
	}).Infof(botName + " started up")

	s.UpdateStatus(0, botKeyphrase) // Set the Discord "playing" status
}
