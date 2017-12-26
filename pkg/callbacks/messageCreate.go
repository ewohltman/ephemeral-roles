package callbacks

import (
	"os"
	"strings"

	"github.com/ewohltman/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

// MessageCreate is the callback function for the MessageCreate event from Discord
func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages from bots
	if m.Author.Bot {
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
		if len(contentTokens) > 2 { // [BOT_KEYWORD] [command] [options] :: !eph log_level debug
			switch strings.ToLower(strings.TrimSpace(contentTokens[1])) {
			case "info":
				// TODO: Reply to bot command info
				// It should provide information about
				// the bot such as what framework it is using and the used
				// version, help commands and, most importantly, who made it.
				//
				// Ignore both your own and other bots' messages. This helps
				// prevent infinite self-loops and potential security exploits.
				// Using a zero width space such as \u200B and \u180E in the
				// beginning of each message also prevents your bot from
				// triggering other bots' commands.
			case "log_level":
				if len(contentTokens) >= 3 {
					logLevel = contentTokens[2]
				}
			default:
				// Silently fail
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
