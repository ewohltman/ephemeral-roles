package members

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

var (
	dgTestBotSession *discordgo.Session
	token            string
	botID            string
	log              = logging.Instance()
)

func TestMain(m *testing.M) {
	var found bool
	token, found = os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}

	botID, found = os.LookupEnv("BOT_ID")
	if !found || botID == "" {
		log.WithField("warn", "BOT_ID not defined in environment variables").
			Warnf("Integration with discordbots.org disabled")
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

	os.Exit(m.Run())
}

func TestMonitor(t *testing.T) {
	check(dgTestBotSession, token, botID)
	update(100, dgTestBotSession, token, botID)
	check(dgTestBotSession, token, botID)
}
