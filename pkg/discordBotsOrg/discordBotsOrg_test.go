package discordBotsOrg

import (
	"os"
	"testing"
)

func TestUpdate(t *testing.T) {
	token := os.Getenv("DISCORDBOTS_ORG_TOKEN")
	botID := os.Getenv("BOT_ID")

	response, err := Update(token, botID, -1)
	if err != nil {
		t.Log("warn: " + err.Error())
		return
	}

	if response != `{"error":"Invalid value for server_count"}` {
		t.Log("warn: discordbots.org integration: unexpected response: " + response)
	}
}
