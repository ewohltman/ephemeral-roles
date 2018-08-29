package guilds

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
)

var (
	dgTestBotSession *discordgo.Session
	token            string
	botID            string
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

func TestHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	_, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected response code from /guilds")
	}
}

func TestMonitor(t *testing.T) {
	check(dgTestBotSession, token, botID)
	update(dgTestBotSession, token, botID)
	check(dgTestBotSession, token, botID)
}
