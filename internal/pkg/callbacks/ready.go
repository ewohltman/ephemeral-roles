package callbacks

import (
	"github.com/bwmarrin/discordgo"
)

const updateBotStatusError = "Unable to update bot status"

// Ready is the callback function for the Ready event from Discord.
func (config *Config) Ready(s *discordgo.Session, event *discordgo.Ready) {
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
		config.Log.WithError(err).Error(updateBotStatusError)
	}
}
