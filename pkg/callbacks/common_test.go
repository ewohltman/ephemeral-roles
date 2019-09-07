package callbacks

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	devGuildID       = "393496906992713746"
	devTextChannelID = "393998570690183168"
	letters          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var dgTestBotSession *discordgo.Session

func init() {
	rand.Seed(time.Now().UnixNano())
}

/*func TestMain(m *testing.M) {
	token, found := os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}

	var err error
	dgTestBotSession, err = discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	err = dgTestBotSession.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}
	defer dgTestBotSession.Close()

	// Wait for asynchronous status to catch up
	// for dgTestBotSession.State.Guilds == nil || !stateContainsGuild(dgTestBotSession.State.Guilds) {
	for !stateContainsGuild(dgTestBotSession.State.Guilds) {
	}

	devGuild, err := dgTestBotSession.State.Guild(devGuildID)
	if err != nil {
		log.WithError(err).Fatalf("Error finding dev guild")
	}

	// for devGuild.Channels == nil || !stateContainsTextChannel(devGuild.Channels) {
	for !stateContainsTextChannel(devGuild.Channels) {
	}

	os.Exit(m.Run())
}

func stateContainsGuild(guilds []*discordgo.Guild) bool {
	for _, guild := range guilds {
		if guild.ID == devGuildID {
			return true
		}
	}
	return false
}

func stateContainsTextChannel(channels []*discordgo.Channel) bool {
	for _, channel := range channels {
		if channel.ID == devTextChannelID {
			return true
		}
	}
	return false
}*/
