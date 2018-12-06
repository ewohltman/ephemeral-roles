package monitor

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor/guilds"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor/members"
)

// Start will launch usage monitors in their own goroutines
func Start(dgBotSession *discordgo.Session) {
	go guilds.Monitor(dgBotSession)
	go members.Monitor(dgBotSession)
}
