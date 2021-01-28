package callbacks

import (
	"github.com/bwmarrin/discordgo"
)

const (
	ready           = "Ready"
	readyEventError = "Unable to process event: " + ready
)

// Ready is the callback function for the Ready event from Discord.
func (handler *Handler) Ready(s *discordgo.Session, event *discordgo.Ready) {
	handler.ReadyCounter.Inc()

	idleSince := 0

	usd := discordgo.UpdateStatusData{
		IdleSince: &idleSince,
		Activities: []*discordgo.Activity{
			{
				Name: handler.BotKeyword,
				Type: discordgo.ActivityTypeListening,
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
