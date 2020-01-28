package environment

import (
	"fmt"
	"os"
)

// Required environment variables.
const (
	Port       = "PORT"
	BotToken   = "BOT_TOKEN"
	BotName    = "BOT_NAME"
	BotKeyword = "BOT_KEYWORD"
	RolePrefix = "ROLE_PREFIX"
	RoleColor  = "ROLE_COLOR_HEX2DEC"
)

// Optional environment variables.
const (
	LogLevel             = "LOG_LEVEL"
	DiscordrusWebHookURL = "DISCORDRUS_WEBHOOK_URL"
	LogTimezoneLocation  = "LOG_TIMEZONE_LOCATION"
	DiscordBotsOrgBotID  = "BOT_ID"
	DiscordBotsOrgToken  = "DISCORDBOTS_ORG_TOKEN" //nolint:gosec // Not a hard-coded credential
)

const (
	undefinedVariable   = "%s not defined in environment variables"
	integrationDisabled = "integration with discordbots.org disabled"
)

// RequiredVariables are the required environment variables.
type RequiredVariables struct {
	Port      string
	BotToken  string
	RoleColor string
}

// OptionalVariables are the optional environment variables.
type OptionalVariables struct {
	discordBotsOrgBotID string
	discordBotsOrgToken string
}

// CheckRequiredVariables checks for required environment variables and returns
// a struct containing those values.
func CheckRequiredVariables() (*RequiredVariables, error) {
	// Check for internal HTTP server port. If not defined, default to 8080
	port, err := lookup(Port)
	if err != nil {
		port = "8080"
	}

	// Check for bot token, we need this to connect to Discord
	botToken, err := lookup(BotToken)
	if err != nil {
		return nil, err
	}

	// Check for other variables. These are not needed now, but are needed in
	// the callbacks
	for _, envVar := range []string{BotName, BotKeyword, RolePrefix, RoleColor} {
		_, err = lookup(envVar)
		if err != nil {
			return nil, err
		}
	}

	return &RequiredVariables{
		Port:     port,
		BotToken: botToken,
	}, nil
}

// CheckOptionalVariables checks for optional environment variables and returns
// a struct containing those values.
func CheckOptionalVariables() (*OptionalVariables, error) {
	// Check for discordbots.org bot ID. We need this for optional
	// discordbots.org integration
	discordBotsOrgBotID, err := lookup(DiscordBotsOrgBotID)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", integrationDisabled, err)
	}

	// Check for discordbots.org token, we need this for optional
	// discordbots.org integration
	discordBotsOrgToken, err := lookup(DiscordBotsOrgToken)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", integrationDisabled, err)
	}

	return &OptionalVariables{
		discordBotsOrgBotID: discordBotsOrgBotID,
		discordBotsOrgToken: discordBotsOrgToken,
	}, nil
}

func lookup(environmentVariable string) (string, error) {
	value, found := os.LookupEnv(environmentVariable)
	if !found || value == "" {
		return "", fmt.Errorf(undefinedVariable, environmentVariable)
	}

	return value, nil
}
