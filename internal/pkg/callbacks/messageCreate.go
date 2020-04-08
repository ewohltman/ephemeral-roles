package callbacks

import (
	"context"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
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
	logLevelParamDebug   = "debug"
	logLevelParamInfo    = "info"
	logLevelParamWarning = "warning"
	logLevelParamError   = "error"
	logLevelParamFatal   = "fatal"
	logLevelParamPanic   = "panic"
)

const (
	messageCreate    = "MessageCreate"
	infoMessageColor = 0xffa500

	logoURLBase = "https://raw.githubusercontent.com/ewohltman/ephemeral-roles"
	logoURLPath = "/master/web/static/Testa_Anatomica-Filippo_Balbi.jpg"
	logoURL     = logoURLBase + logoURLPath

	logLevelChange = "Logging level changed"
)

// MessageCreate is the callback function for the MessageCreate event from Discord
func (config *Config) MessageCreate(session *discordgo.Session, mc *discordgo.MessageCreate) {
	// Increment the total number of MessageCreate events
	config.MessageCreateCounter.Inc()

	span := config.JaegerTracer.StartSpan(messageCreate)
	defer span.Finish()

	ctx, cancelCtx := context.WithTimeout(context.Background(), config.ContextTimeout)
	defer cancelCtx()

	ctx = opentracing.ContextWithSpan(ctx, span)

	// Ignore all messages from bots
	if mc.Author.Bot {
		return
	}

	// [BOT_KEYWORD] [command] [options] :: "!eph" "log_level" "debug"
	contentTokens := strings.Split(strings.TrimSpace(mc.Content), " ")
	if len(contentTokens) < numTokensMinimum {
		return
	}

	// Check if the message starts with our keyword
	if contentTokens[0] != config.BotKeyword {
		return
	}

	// Find the guild
	guild, err := lookupGuild(ctx, session, mc.GuildID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to find guild")
		return
	}

	// Find the channel
	channel, err := session.State.Channel(mc.ChannelID)
	if err != nil {
		config.Log.WithError(err).Debugf("Unable to find channel")
		return
	}

	config.Log.WithFields(logrus.Fields{
		"author":        mc.Author.Username,
		"content":       mc.Content,
		"contentTokens": contentTokens,
		"channel":       channel.Name,
		"guild":         guild.Name,
	}).Debugf("New message")

	config.parseMessage(session, mc.ChannelID, contentTokens)
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
		level := strings.ToLower(contentTokens[2])

		logFields := logrus.Fields{logLevelCommand: level}

		switch level {
		case logLevelParamDebug:
			config.updateLogLevel(level)
			config.Log.WithFields(logFields).Debugf(logLevelChange)
		case logLevelParamInfo:
			config.updateLogLevel(level)
			config.Log.WithFields(logFields).Infof(logLevelChange)
		case logLevelParamWarning:
			config.updateLogLevel(level)
			config.Log.WithFields(logFields).Warnf(logLevelChange)
		case logLevelParamError:
			config.updateLogLevel(level)
			config.Log.WithFields(logFields).Errorf(logLevelChange)
		case logLevelParamFatal:
			config.updateLogLevel(level)
		case logLevelParamPanic:
			config.updateLogLevel(level)
		}
	}
}

func (config *Config) updateLogLevel(level string) {
	config.Log.UpdateLevel(level)
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
