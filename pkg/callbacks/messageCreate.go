package callbacks

import (
	"os"
	"strings"

	"github.com/ewohltman/ephemeral-roles/pkg/environment"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

const (
	logoURL = "https://raw.githubusercontent.com/ewohltman/ephemeral-roles" +
		"/master/web/static/logo_Testa_anatomica_(1854)_-_Filippo_Balbi.jpg"
	logLevelChange = "Logging level changed"
)

// MessageCreate is the callback function for the MessageCreate event from Discord
func (config *Config) MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Increment the total number of MessageCreate events
	config.MessageCreateCounter.Inc()

	// Ignore all messages from bots
	if m.Author.Bot {
		return
	}

	// [BOT_KEYWORD] [command] [options] :: "!eph" "log_level" "debug"
	contentTokens := strings.Split(strings.TrimSpace(m.Content), " ")
	if len(contentTokens) < 1 {
		return
	}

	// Check if the message starts with our keyword
	if contentTokens[0] != config.BotKeyword {
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

	config.Log.WithFields(logrus.Fields{
		"author":        m.Author.Username,
		"content":       m.Content,
		"contentTokens": contentTokens,
		"channel":       c.Name,
		"guild":         g.Name,
	}).Debugf("New message")

	config.parseMessage(s, m.ChannelID, contentTokens)
}

func (config *Config) parseMessage(s *discordgo.Session, channelID string, contentTokens []string) {
	if len(contentTokens) < 2 {
		config.handleInfo(s, channelID)
		return
	}

	switch strings.ToLower(contentTokens[1]) {
	case "info":
		config.handleInfo(s, channelID)
	case "log_level":
		config.handleLogLevel(contentTokens)
	default: // Do nothing for unrecognized command
	}
}

func (config *Config) handleInfo(s *discordgo.Session, channelID string) {
	_, err := s.ChannelMessageSendEmbed(channelID, infoMessage())
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to send info message")
	}
}

func (config *Config) handleLogLevel(contentTokens []string) {
	if len(contentTokens) >= 3 {
		levelOpt := strings.ToLower(contentTokens[2])

		logFields := logrus.Fields{"log_level": levelOpt}

		switch levelOpt {
		case "debug":
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Debugf(logLevelChange)
		case "info":
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Infof(logLevelChange)
		case "warn":
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Warnf(logLevelChange)
		case "error":
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Errorf(logLevelChange)
		case "fatal":
			config.updateLogLevel(levelOpt)
		case "panic":
			config.updateLogLevel(levelOpt)
		}
	}
}

func (config *Config) updateLogLevel(levelOpt string) {
	err := os.Setenv("LOG_LEVEL", levelOpt)
	if err != nil {
		config.Log.
			WithError(err).
			Warnf("Unable to set %s environment variable", environment.LogLevel)
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
			URL: logoURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "About",
				Value:  "Ephemeral Roles is a Discord bot designed to assign roles based upon voice channel member presence.",
				Inline: false,
			},
			{
				Name:   "Author",
				Value:  "Ephemeral Roles was created by ewohltman: https://github.com/ewohltman",
				Inline: false,
			},
			{
				Name:   "Library",
				Value:  "Ephemeral Roles uses the `discordgo` library by bwmarrin: https://github.com/bwmarrin/discordgo",
				Inline: false,
			},
		},
	}
}
