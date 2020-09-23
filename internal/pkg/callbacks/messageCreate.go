package callbacks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

// Supported commands
const (
	InfoCommand     = "info"
	LogLevelCommand = "log_level"
)

// Supported command parameters
const (
	LogLevelParamDebug   = "debug"
	LogLevelParamInfo    = "info"
	LogLevelParamWarning = "warning"
	LogLevelParamError   = "error"
	LogLevelParamFatal   = "fatal"
	LogLevelParamPanic   = "panic"
)

// Content token parsing
const (
	numTokensMinimum               = 1
	numTokensWithCommand           = 2
	numTokensWithCommandParameters = 3
)

const (
	messageCreate    = "MessageCreate"
	infoMessageColor = 0xffa500

	logoURLBase = "https://raw.githubusercontent.com/ewohltman/ephemeral-roles"
	logoURLPath = "/master/web/static/Testa_Anatomica-Filippo_Balbi.jpg"
	logoURL     = logoURLBase + logoURLPath

	logLevelChange = "Logging level changed"

	messageCreateEventError = "Unable to process event: " + messageCreate
)

// MessageCreate is the callback function for the MessageCreate event from Discord.
func (config *Config) MessageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	config.MessageCreateCounter.Inc()

	if message.Author.Bot {
		return
	}

	// [BOT_KEYWORD] [command] [options] :: "!eph" "log_level" "debug"
	contentTokens := strings.Split(strings.TrimSpace(message.Content), " ")
	if len(contentTokens) < numTokensMinimum {
		return
	}

	if contentTokens[0] != config.BotKeyword {
		return
	}

	err := config.parseMessage(session, contentTokens, message.ChannelID)
	if err != nil {
		config.Log.WithError(err).Error(messageCreateEventError)
	}
}

func (config *Config) parseMessage(s *discordgo.Session, contentTokens []string, channelID string) error {
	if len(contentTokens) < numTokensWithCommand {
		err := config.handleInfo(s, channelID)
		if err != nil {
			return err
		}

		return nil
	}

	switch strings.ToLower(contentTokens[1]) {
	case InfoCommand:
		err := config.handleInfo(s, channelID)
		if err != nil {
			return err
		}
	case LogLevelCommand:
		config.handleLogLevel(contentTokens)
	}

	return nil
}

func (config *Config) handleInfo(s *discordgo.Session, channelID string) error {
	_, err := s.ChannelMessageSendEmbed(channelID, infoMessage())
	if err != nil {
		return fmt.Errorf("error sending info message: %w", err)
	}

	return nil
}

func (config *Config) handleLogLevel(contentTokens []string) {
	if len(contentTokens) >= numTokensWithCommandParameters {
		logLevel := strings.ToLower(contentTokens[2])

		logFields := logrus.Fields{LogLevelCommand: logLevel}

		config.updateLogLevel(logLevel)

		switch logLevel {
		case LogLevelParamDebug:
			config.Log.WithFields(logFields).Debugf(logLevelChange)
		case LogLevelParamInfo:
			config.Log.WithFields(logFields).Infof(logLevelChange)
		case LogLevelParamWarning:
			config.Log.WithFields(logFields).Warnf(logLevelChange)
		case LogLevelParamError:
			config.Log.WithFields(logFields).Errorf(logLevelChange)
		}
	}
}

func (config *Config) updateLogLevel(logLevel string) {
	config.Log.UpdateLevel(logLevel)
	config.Log.UpdateDiscordrus()
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
