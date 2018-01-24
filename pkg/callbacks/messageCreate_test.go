package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

const devChannel = "393998570690183168"

func TestMessageCreate(t *testing.T) {
	dgTestBotSession.ChannelMessageSendComplex(
		devChannel,
		&discordgo.MessageSend{
			Content: "AUTOMATED TESTING",
		},
	)

	content := ""

	// bot
	botMsg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "AUTOMATED TEST BOT USER",
				Bot:      true,
			},
		},
	}
	MessageCreate(dgTestBotSession, botMsg)

	// non keyphrase
	content = "this should not show up!"
	nonKeyphraseMsg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "AUTOMATED TEST USER",
				Bot:      false,
			},
			ChannelID: devChannel,
			Content:   content,
		},
	}
	MessageCreate(dgTestBotSession, nonKeyphraseMsg)

	// keyphrase
	content = BOTKEYWORD + " AUTOMATED TEST"
	keyphraseMsg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "AUTOMATED TEST USER",
				Bot:      false,
			},
			ChannelID: devChannel,
			Content:   content,
		},
	}
	MessageCreate(dgTestBotSession, keyphraseMsg)
}
