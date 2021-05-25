package callbacks_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/pkg/mockconstants"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

func TestHandler_MessageCreate(t *testing.T) {
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

	log := mock.NewLogger()

	handler := &callbacks.Handler{
		Log:                  log,
		BotName:              "testBot",
		BotKeyword:           "testKeyword",
		RolePrefix:           "{eph}",
		JaegerTracer:         jaegerTracer,
		ContextTimeout:       time.Second,
		MessageCreateCounter: monitor.MessageCreateCounter(&monitor.Config{Log: log}),
	}

	originalLogLevel := log.Level.String()

	sendBotMessage(session, handler)

	tests := []string{
		"",                 // no words
		"ixnay",            // no keyword
		handler.BotKeyword, // only keyword
		fmt.Sprintf("%s %s", handler.BotKeyword, "ixnay"), // keyword, unrecognized command
		fmt.Sprintf("%s %s", handler.BotKeyword, callbacks.InfoCommand),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, callbacks.LogLevelParamDebug),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, callbacks.LogLevelParamInfo),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, callbacks.LogLevelParamWarning),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, callbacks.LogLevelParamError),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, callbacks.LogLevelParamFatal),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, callbacks.LogLevelParamPanic),
		fmt.Sprintf("%s %s %s", handler.BotKeyword, callbacks.LogLevelCommand, originalLogLevel),
	}

	for _, test := range tests {
		sendMessage(session, handler, test)
	}
}

func sendBotMessage(session *discordgo.Session, handler *callbacks.Handler) {
	handler.MessageCreate(session, &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: handler.BotName,
				Bot:      true,
			},
			GuildID:   mockconstants.TestGuild,
			ChannelID: mockconstants.TestChannel,
			Content:   "",
		},
	})
}

func sendMessage(s *discordgo.Session, handler *callbacks.Handler, message string) {
	handler.MessageCreate(s, &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: handler.BotName,
				Bot:      false,
			},
			GuildID:   mockconstants.TestGuild,
			ChannelID: mockconstants.TestChannel,
			Content:   message,
		},
	})
}
