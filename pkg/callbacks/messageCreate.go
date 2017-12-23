package callbacks

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

// MessageCreate is the callback function for the "Message create" event from Discord
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message starts with our keyword
	if strings.HasPrefix(m.Content, BOT_KEYWORD+" ") {
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

		contentTokens := strings.Split(strings.TrimSpace(m.Content), " ")

		log.WithFields(logrus.Fields{
			"author":        m.Author.Username,
			"channel":       c.Name,
			"guild":         g.Name,
			"content":       m.Content,
			"contentTokens": contentTokens,
		}).Debugf("New message")

		logLevel := ""
		if len(contentTokens) == 3 { // [BOT_KEYWORD] [modifyThis] [toThis] :: !eph log_level debug
			switch strings.ToLower(strings.TrimSpace(contentTokens[1])) {
			case "log_level":
				logLevel = contentTokens[2]
			}

			if logLevel != "" {
				os.Setenv("LOG_LEVEL", logLevel)

				logging.Reinitialize()
			}

			log.WithFields(logrus.Fields{
				"author":    m.Author.Username,
				"channel":   c.Name,
				"guild":     g.Name,
				"content":   m.Content,
				"LOG_LEVEL": logLevel,
			}).Infof("Logging level changed")
		}
	}
}
