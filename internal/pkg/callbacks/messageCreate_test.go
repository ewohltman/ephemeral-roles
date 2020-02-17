package callbacks

import (
	"fmt"
	"testing"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestConfig_MessageCreate(t *testing.T) {
	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	log := mock.NewLogger()

	monitorConfig := &monitor.Config{
		Log: log,
	}

	config := &Config{
		Log:                     log,
		BotName:                 "testBot",
		BotKeyword:              "testKeyword",
		RolePrefix:              "testRolePrefix",
		ReadyCounter:            nil,
		MessageCreateCounter:    monitorConfig.MessageCreateCounter(),
		VoiceStateUpdateCounter: nil,
	}

	originalLogLevel := log.Level.String()

	// message from a bot
	sendBotMessage(session, config)

	tests := []string{
		"ixnay",           // no keyword
		config.BotKeyword, // only keyword
		fmt.Sprintf("%s %s", config.BotKeyword, "ixnay"), // keyword, unrecognized command
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, logLevelParamDebug),
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, logLevelParamInfo),
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, logLevelParamWarning),
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, logLevelParamError),
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, logLevelParamFatal),
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, logLevelParamPanic),
		fmt.Sprintf("%s %s %s", config.BotKeyword, logLevelCommand, originalLogLevel),
	}

	for _, test := range tests {
		sendMessage(session, config, test)
	}
}

func sendBotMessage(s *discordgo.Session, config *Config) {
	config.MessageCreate(
		s,
		&discordgo.MessageCreate{
			Message: &discordgo.Message{
				Author: &discordgo.User{
					Username: config.BotName,
					Bot:      true,
				},
				GuildID:   mock.TestGuild,
				ChannelID: mock.TestChannel,
				Content:   "",
			},
		},
	)
}

func sendMessage(s *discordgo.Session, config *Config, message string) {
	config.MessageCreate(
		s,
		&discordgo.MessageCreate{
			Message: &discordgo.Message{
				Author: &discordgo.User{
					Username: config.BotName,
					Bot:      false,
				},
				GuildID:   mock.TestGuild,
				ChannelID: mock.TestChannel,
				Content:   message,
			},
		},
	)
}
