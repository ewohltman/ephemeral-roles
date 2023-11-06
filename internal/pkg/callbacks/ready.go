package callbacks

import (
	"github.com/bwmarrin/discordgo"
)

const readyEventError = unableToProcessEvent + "Ready"

// Ready is the callback function for the Ready event from Discord.
func (handler *Handler) Ready(s *discordgo.Session, _ *discordgo.Ready) {
	handler.ReadyCounter.Inc()

	idleSince := 0

	usd := discordgo.UpdateStatusData{
		IdleSince: &idleSince,
		Activities: []*discordgo.Activity{
			{
				Name: "voice channels",
				Type: discordgo.ActivityTypeWatching,
			},
		},
		AFK:    false,
		Status: "online",
	}

	err := s.UpdateStatusComplex(usd)
	if err != nil {
		handler.Log.WithError(err).Error(readyEventError)
	}
}
