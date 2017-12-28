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

	// Check if the message starts with our keyword
	if strings.HasPrefix(m.Content, BOTKEYWORD+" ") {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.WithError(err).Debugf("Unable to find channel")

			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			log.WithError(err).Debugf("Unable to find guild")

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

		if len(contentTokens) > 2 { // [BOT_KEYWORD] [command] [options] :: !eph log_level debug
			logFields := logrus.Fields{
				"author":  m.Author.Username,
				"channel": c.Name,
				"guild":   g.Name,
				"content": m.Content,
			}

			switch strings.ToLower(contentTokens[1]) {
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
					levelOpt := strings.ToLower(contentTokens[2])

					logFields["LOG_LEVEL"] = levelOpt

					switch levelOpt {
					case "debug":
						updateLogLevel(levelOpt)
						log.WithFields(logFields).Debugf("Logging level changed")
					case "info":
						updateLogLevel(levelOpt)
						log.WithFields(logFields).Infof("Logging level changed")
					case "warn":
						updateLogLevel(levelOpt)
						log.WithFields(logFields).Warnf("Logging level changed")
					case "error":
						updateLogLevel(levelOpt)
						log.WithFields(logFields).Errorf("Logging level changed")
					case "fatal":
						updateLogLevel(levelOpt)
					case "panic":
						updateLogLevel(levelOpt)
					}
				}
			default:
				// Silently fail for unrecognized command
			}
		}
	}
}

func updateLogLevel(levelOpt string) {
	os.Setenv("LOG_LEVEL", levelOpt)

	logging.Reinitialize()
}
