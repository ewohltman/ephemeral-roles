package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestMessageCreate(t *testing.T) {
	botKeyword := "!eph "

	testSession := &discordgo.Session{}

	// message from a bot
	sendBotMessage(testSession)

	// non keyphrase message
	sendMessage(testSession, "ixnay")

	// keyphrase message, unrecognized command
	sendMessage(testSession, botKeyword+"ixnay")

	// keyphrase message, unrecognized command
	sendMessage(testSession, botKeyword+"AUTOMATED TEST")

	// keyphrase info
	sendMessage(testSession, botKeyword+"info")

	// log_level debug
	sendMessage(testSession, botKeyword+"log_level debug")

	// log_level info
	sendMessage(testSession, botKeyword+"log_level info")

	// log_level warn
	sendMessage(testSession, botKeyword+"log_level warn")

	// log_level error
	sendMessage(testSession, botKeyword+"log_level error")

	// log_level fatal
	sendMessage(testSession, botKeyword+"log_level fatal")

	// log_level panic
	sendMessage(testSession, botKeyword+"log_level panic")

	// log_level info
	sendMessage(testSession, botKeyword+"log_level info")
}

func sendBotMessage(s *discordgo.Session) {
	botMsg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "",
				Bot:      true,
			},
			GuildID:   "",
			ChannelID: "",
			Content:   "",
		},
	}

	MessageCreate()(s, botMsg)
}

func sendMessage(s *discordgo.Session, message string) {
	msg := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				Username: "",
				Bot:      false,
			},
			GuildID:   "",
			ChannelID: "",
			Content:   message,
		},
	}

	MessageCreate()(s, msg)
}
