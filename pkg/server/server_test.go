package server

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
)

var (
	token            string
	botID            string
	dgTestBotSession *discordgo.Session
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

func TestNew(t *testing.T) {
	testServer := New("8080")
	if testServer == nil {
		t.Errorf("Failed creating new internal HTTP server")
	}
}
