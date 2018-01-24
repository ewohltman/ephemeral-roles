package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestMessageCreate(t *testing.T) {
	testBotMessage := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "TEST BOT USER",
				Bot:      true,
			},
		},
	}

	MessageCreate(dgTestBotSession, testBotMessage)

	testNonKeyphraseMessage := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "TEST USER",
				Bot:      false,
			},
			Content: "abcd",
		},
	}

	MessageCreate(dgTestBotSession, testNonKeyphraseMessage)
}
