package callbacks

import (
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/environment"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

// Content token parsing
const (
	numTokensMinimum               = 1
	numTokensWithCommand           = 2
	numTokensWithCommandParameters = 3
)

// Supported commands
const (
	infoCommand     = "info"
	logLevelCommand = "log_level"
)

// Supported command parameters
const (
	logLevelParamDebug = "debug"
	logLevelParamInfo  = "info"
	logLevelParamWarn  = "warn"
	logLevelParamError = "error"
	logLevelParamFatal = "fatal"
	logLevelParamPanic = "panic"
)

const (
	infoMessageColor = 0xffa500

	logoURLBase = "https://raw.githubusercontent.com/ewohltman/ephemeral-roles"
	logoURLPath = "/master/web/static/logo_Testa_anatomica_(1854)_-_Filippo_Balbi.jpg"
	logoURL     = logoURLBase + logoURLPath

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
	if len(contentTokens) < numTokensMinimum {
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
	if len(contentTokens) < numTokensWithCommand {
		config.handleInfo(s, channelID)
		return
	}

	switch strings.ToLower(contentTokens[1]) {
	case infoCommand:
		config.handleInfo(s, channelID)
	case logLevelCommand:
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
	if len(contentTokens) >= numTokensWithCommandParameters {
		levelOpt := strings.ToLower(contentTokens[2])

		logFields := logrus.Fields{"log_level": levelOpt}

		switch levelOpt {
		case logLevelParamDebug:
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Debugf(logLevelChange)
		case logLevelParamInfo:
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Infof(logLevelChange)
		case logLevelParamWarn:
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Warnf(logLevelChange)
		case logLevelParamError:
			config.updateLogLevel(levelOpt)
			config.Log.WithFields(logFields).Errorf(logLevelChange)
		case logLevelParamFatal:
			config.updateLogLevel(levelOpt)
		case logLevelParamPanic:
			config.updateLogLevel(levelOpt)
		}
	}
}

func (config *Config) updateLogLevel(levelOpt string) {
	err := os.Setenv("LOG_LEVEL", levelOpt)
	if err != nil {
		config.Log.WithError(err).Warnf("Unable to set %s environment variable", environment.LogLevel)
		return
	}

	logging.UpdateLevel(config.Log)
}

func infoMessage() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		URL:   "https://github.com/ewohltman/ephemeral-roles",
		Title: "Ephemeral Roles",
		Color: infoMessageColor,
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
