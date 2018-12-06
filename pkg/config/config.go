package config

import (
	"os"

	"github.com/pkg/errors"
)

// CheckRequired will check the environment for required values
func CheckRequired() (token string, port string, err error) {
	var found bool

	// Check for BOT_TOKEN, we need this to connect to Discord
	token, found = os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		err = errors.New("BOT_TOKEN not defined in environment variables")
		return
	}

	// Check for PORT, we need this to for our HTTP server in our container
	port, found = os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}

	// Check for strings from slice, these are not needed now, but are needed in the callbacks
	for _, envVar := range []string{"BOT_NAME", "BOT_KEYWORD", "ROLE_PREFIX"} {
		value, found := os.LookupEnv(envVar)
		if !found || value == "" {
			err = errors.New("%s not defined in environment variables" + envVar)
			return
		}
	}

	return
}

// CheckDiscordBotsOrg will check the environment for optional values
func CheckDiscordBotsOrg() (botID string, discordBotsOrgToken string, err error) {
	var found bool

	// Check for BOT_ID and DISCORDBOTS_ORG_TOKEN, we need these for optional discordbots.org integration
	botID, found = os.LookupEnv("BOT_ID")
	if !found || botID == "" {
		err = errors.New("integration with discordbots.org disabled: BOT_ID not defined in environment variables")
		return
	}

	discordBotsOrgToken, found = os.LookupEnv("DISCORDBOTS_ORG_TOKEN")
	if !found || discordBotsOrgToken == "" {
		err = errors.New("integration with discordbots.org disabled: DISCORDBOTS_ORG_TOKEN not defined in environment variables")
		return
	}

	return
}
