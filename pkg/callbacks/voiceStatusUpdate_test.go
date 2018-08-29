package callbacks

import (
	"math/rand"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestVoiceStateUpdate(t *testing.T) {
	devVoiceChannel1, err := dgTestBotSession.GuildChannelCreate(devGuildID, randString(5), discordgo.ChannelTypeGuildVoice)
	if err != nil {
		t.Error(err)
	}
	defer dgTestBotSession.ChannelDelete(devVoiceChannel1.ID)

	devVoiceChannel2, err := dgTestBotSession.GuildChannelCreate(devGuildID, randString(5), discordgo.ChannelTypeGuildVoice)
	if err != nil {
		t.Error(err)
	}
	defer dgTestBotSession.ChannelDelete(devVoiceChannel2.ID)

	// connect
	sendUpdate(devVoiceChannel1.ID)

	//change
	sendUpdate(devVoiceChannel2.ID)

	// disconnect
	sendUpdate("")

	// reconnect
	sendUpdate(devVoiceChannel1.ID)

	// disconnect
	sendUpdate("")
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}

func sendUpdate(channelID string) {
	update := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    dgTestBotSession.State.User.ID,
			GuildID:   devGuildID,
			ChannelID: channelID,
		},
	}
	VoiceStateUpdate(dgTestBotSession, update)
}
