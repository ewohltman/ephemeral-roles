package callbacks

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// MessageCreate is the callback function for the "Message create" event from Discord
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is BotKeyword
	if strings.HasPrefix(m.Content, botKeyphrase) {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.WithError(err).Errorf("Unable to find channel")

			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			log.WithError(err).Errorf("Unable to find guild")

			return
		}

		log.WithFields(logrus.Fields{
			"author":  m.Author.Username,
			"channel": c.Name,
			"guild":   g.Name,
		}).Debugf("New message")
	}
}
