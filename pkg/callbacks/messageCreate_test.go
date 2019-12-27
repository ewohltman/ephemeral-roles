package callbacks

import (
	"fmt"
	"testing"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"

	"github.com/bwmarrin/discordgo"
)

func TestConfig_MessageCreate(t *testing.T) {
	session, err := mock.Session()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

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
		"ixnay", // no keyword
		fmt.Sprintf("%s %s", config.BotKeyword, "ixnay"), // keyword, unrecognized command
		fmt.Sprintf("%s %s", config.BotKeyword, "info"),  // keyword, incomplete command
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level debug"),
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level info"),
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level warn"),
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level error"),
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level fatal"),
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level panic"),
		fmt.Sprintf("%s %s", config.BotKeyword, "log_level "+originalLogLevel),
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
				GuildID:   "testGuild",
				ChannelID: "testChannel",
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
				GuildID:   "testGuild",
				ChannelID: "testChannel",
				Content:   message,
			},
		},
	)
}
