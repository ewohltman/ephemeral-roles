package server

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
)

var dgTestBotSession *discordgo.Session

func TestMain(m *testing.M) {
	serverTest = true

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

	// Initialize
	guildsUpdate(dgTestBotSession, token, "")

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	testServer := New("8080")
	if testServer == nil {
		t.Errorf("Failed creating new internal HTTP server")
	}
}

func TestMonitorGuildsUpdate(t *testing.T) {
	token, found := os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}

	// No change
	isc.mu.Lock()
	isc.numGuilds = len(dgTestBotSession.State.Guilds)
	isc.mu.Unlock()
	MonitorGuildsUpdate(dgTestBotSession, token, "")

	// Added to guild
	isc.mu.Lock()
	isc.numGuilds = len(dgTestBotSession.State.Guilds) - 1
	isc.mu.Unlock()
	MonitorGuildsUpdate(dgTestBotSession, token, "")

	// Removed from guild
	isc.mu.Lock()
	isc.numGuilds = len(dgTestBotSession.State.Guilds) + 1
	isc.mu.Unlock()
	MonitorGuildsUpdate(dgTestBotSession, token, "")
}
