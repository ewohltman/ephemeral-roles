package callbacks

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

// MessageCreate is the callback function for the MessageCreate event from Discord
func (config *Config) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Increment the total number of MessageCreate events
	config.MessageCreateCounter.Inc()

	// Ignore all messages from bots
	if m.Author.Bot {
		return
	}

	// Check if the message starts with our keyword
	if !strings.HasPrefix(m.Content, config.BotKeyword) {
		return
	}

	// [BOT_KEYWORD] [command] [options] :: "!eph" "log_level" "debug"
	contentTokens := strings.Split(strings.TrimSpace(m.Content), " ")
	if len(contentTokens) < 2 {
		return
	}

	// Find the channel
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to find channel")
		return
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to find guild")
		return
	}

	logFields := logrus.Fields{
		"author":        m.Author.Username,
		"content":       m.Content,
		"contentTokens": contentTokens,
		"channel":       c.Name,
		"guild":         g.Name,
	}

	config.Log.WithFields(logFields).Debugf("New message")

	switch strings.ToLower(contentTokens[1]) {
	case "info":
		_, err := s.ChannelMessageSendEmbed(m.ChannelID, infoMessage())
		if err != nil {
			config.Log.WithError(err).Debugf("Unable to send message")
			return
		}
	case "log_level":
		if len(contentTokens) >= 3 {
			levelOpt := strings.ToLower(contentTokens[2])

			logFields["log_level"] = levelOpt

			switch levelOpt {
			case "debug":
				config.updateLogLevel(levelOpt)
				config.Log.WithFields(logFields).Debugf("Logging level changed")
			case "info":
				config.updateLogLevel(levelOpt)
				config.Log.WithFields(logFields).Infof("Logging level changed")
			case "warn":
				config.updateLogLevel(levelOpt)
				config.Log.WithFields(logFields).Warnf("Logging level changed")
			case "error":
				config.updateLogLevel(levelOpt)
				config.Log.WithFields(logFields).Errorf("Logging level changed")
			case "fatal":
				config.updateLogLevel(levelOpt)
			case "panic":
				config.updateLogLevel(levelOpt)
			}
		}
	default:
		// Silently fail for unrecognized command
	}
}

func (config *Config) updateLogLevel(levelOpt string) {
	err := os.Setenv("LOG_LEVEL", levelOpt)
	if err != nil {
		config.Log.WithError(err).Warn("Unable to set LOG_LEVEL environment variable")
		return
	}

	logging.UpdateLevel(config.Log)
}

func infoMessage() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		URL:   "https://github.com/ewohltman/ephemeral-roles",
		Title: "Ephemeral Roles",
		Color: 0xffa500,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Made using the discordgo library",
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/logo_Testa_anatomica_(1854)_-_Filippo_Balbi.jpg",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "About",
				Value:  "Ephemeral Roles is a discord bot designed to assign roles based upon voice channel member presence",
				Inline: false,
			},
			{
				Name:   "Author",
				Value:  "Ephemeral Roles is created by ewohltman",
				Inline: false,
			},
			{
				Name:   "Library",
				Value:  "Ephemeral Roles uses the discordgo library by bwmarrin",
				Inline: false,
			},
		},
	}
}
