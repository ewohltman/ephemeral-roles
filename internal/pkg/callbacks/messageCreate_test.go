package callbacks

import (
	"fmt"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

func TestConfig_MessageCreate(t *testing.T) {
	jaegerTracer, jaegerCloser, err := tracer.New("test")
	if err != nil {
		t.Fatalf("Error creating Jaeger tracer: %s", err)
	}

	defer func() {
		closeErr := jaegerCloser.Close()
		if closeErr != nil {
			t.Errorf("Error closing Jaeger tracer: %s", err)
		}
	}()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	session.Client = http.NewClient(
		session.Client.Transport,
		jaegerTracer,
		"test-0",
	)

	defer mock.SessionClose(t, session)

	log := mock.NewLogger()

	monitorConfig := &monitor.Config{
		Log: log,
	}

	config := &Config{
		Log:                  log,
		BotName:              "testBot",
		BotKeyword:           "testKeyword",
		RolePrefix:           "{eph}",
		JaegerTracer:         jaegerTracer,
		ContextTimeout:       time.Second,
		MessageCreateCounter: monitorConfig.MessageCreateCounter(),
	}

	originalLogLevel := log.Level.String()

	sendBotMessage(session, config)

	tests := []string{
		"ixnay",           // no keyword
		config.BotKeyword, // only keyword
		fmt.Sprintf("%s %s", config.BotKeyword, "ixnay"), // keyword, unrecognized command
		fmt.Sprintf("%s %s", config.BotKeyword, infoCommand),
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

func sendBotMessage(session *discordgo.Session, config *Config) {
	config.MessageCreate(
		session,
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
