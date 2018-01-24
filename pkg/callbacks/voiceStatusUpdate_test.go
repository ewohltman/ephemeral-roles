package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

const (
	devGuildID         = "393496906992713746"
	devVoiceChannelID1 = "393496908062130203"
	devVoiceChannelID2 = "393497797749375017"
)

func TestVoiceStateUpdate(t *testing.T) {
	connect := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    dgTestBotSession.State.User.ID,
			GuildID:   devGuildID,
			ChannelID: devVoiceChannelID1,
		},
	}
	VoiceStateUpdate(dgTestBotSession, connect)

	change := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    dgTestBotSession.State.User.ID,
			GuildID:   devGuildID,
			ChannelID: devVoiceChannelID2,
		},
	}
	VoiceStateUpdate(dgTestBotSession, change)

	disconnect := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    dgTestBotSession.State.User.ID,
			GuildID:   devGuildID,
			ChannelID: "",
		},
	}
	VoiceStateUpdate(dgTestBotSession, disconnect)
}
