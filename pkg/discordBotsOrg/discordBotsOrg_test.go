package discordBotsOrg

import (
	"os"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	token := os.Getenv("DISCORDBOTS_ORG_TOKEN")
	botID := os.Getenv("BOT_ID")

	err := Update(token, botID, -1)
	if err != nil {
		if !strings.HasSuffix(err.Error(), `{"error":"Forbidden"}`) {
			t.Log(err)
		}
	}
}
