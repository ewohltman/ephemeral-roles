package callbacks

import (
	"math/rand"
	"os"
	"testing"
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

func TestMain(m *testing.M) {
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
	for dgTestBotSession.State.Guilds == nil || len(dgTestBotSession.State.Guilds) == 0 {
	}

	for dgTestBotSession.State.Guilds[0].Channels == nil || len(dgTestBotSession.State.Guilds[0].Channels) == 0 {
	}

	os.Exit(m.Run())
}
