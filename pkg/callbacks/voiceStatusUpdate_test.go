package callbacks

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestVoiceStateUpdate(t *testing.T) {
	/*devVoiceChannel1, err := dgTestBotSession.GuildChannelCreate(devGuildID, randString(5), discordgo.ChannelTypeGuildVoice)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer dgTestBotSession.ChannelDelete(devVoiceChannel1.ID)

	devVoiceChannel2, err := dgTestBotSession.GuildChannelCreate(devGuildID, randString(5), discordgo.ChannelTypeGuildVoice)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer dgTestBotSession.ChannelDelete(devVoiceChannel2.ID)*/

	testSession := &discordgo.Session{
		State:        discordgo.NewState(),
		StateEnabled: true,
		Ratelimiter:  discordgo.NewRatelimiter(),
	}

	testUser := &discordgo.User{
		ID:       "testUser",
		Username: "Test User",
	}

	err := testSession.State.GuildAdd(
		&discordgo.Guild{
			ID: "testGuild",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = testSession.State.MemberAdd(
		&discordgo.Member{
			User:    testUser,
			Nick:    "Test User",
			GuildID: "testGuild",
		},
	)

	err = testSession.State.ChannelAdd(
		&discordgo.Channel{
			ID:      "testChannel",
			Name:    "Channel Name",
			GuildID: "testGuild",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	sendUpdate(testSession)

	/*// connect
	sendUpdate(testSession, "channel1")

	//change
	sendUpdate(testSession, devVoiceChannel2.ID)

	// disconnect
	sendUpdate(testSession, "")

	// reconnect
	sendUpdate(testSession, devVoiceChannel1.ID)

	// reconnect same channel
	sendUpdate(testSession, devVoiceChannel1.ID)

	// disconnect
	sendUpdate(testSession, "")*/
}

/*func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}*/

func sendUpdate(s *discordgo.Session) {
	update := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "testUser",
			GuildID:   "testGuild",
			ChannelID: "testChannel",
		},
	}
	VoiceStateUpdate(s, update)
}
