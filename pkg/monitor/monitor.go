package monitor

import (
	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor/guilds"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor/members"
)

// Start will launch usage monitors in their own goroutines
func Start(dgBotSession *discordgo.Session, token string, botID string) {
	go guilds.Monitor(dgBotSession, token, botID)
	go members.Monitor(dgBotSession)
}
