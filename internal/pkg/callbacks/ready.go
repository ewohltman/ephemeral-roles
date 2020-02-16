package callbacks

import (
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Ready is the callback function for the Ready event from Discord
func (config *Config) Ready(s *discordgo.Session, event *discordgo.Ready) {
	// Increment the total number of Ready events
	config.ReadyCounter.Inc()

	idleSince := 0

	usd := discordgo.UpdateStatusData{
		IdleSince: &idleSince,
		Game: &discordgo.Game{
			Name: config.BotKeyword,
			Type: discordgo.GameTypeWatching,
		},
		AFK:    false,
		Status: "online",
	}

	err := s.UpdateStatusComplex(usd)
	if err != nil {
		config.Log.WithError(err).WithFields(logrus.Fields{
			"UpdateStatusData": usd,
		}).Errorf("Error updating complex status")
	}
}
