package callbacks

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
)

var dgTestBotSession *discordgo.Session

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

	for len(dgTestBotSession.State.Guilds) == 0 {
	}

	os.Exit(m.Run())
}
