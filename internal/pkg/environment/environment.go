// Package environment provides implementations for parsing expected
// environment variables.
package environment

import (
	"fmt"
	"os"
)

// Required environment variables.
const (
	BotToken = "BOT_TOKEN"
)

// Optional environment variables.
const (
	LogLevel             = "LOG_LEVEL"
	LogTimezoneLocation  = "LOG_TIMEZONE_LOCATION"
	Port                 = "PORT"
	BotName              = "BOT_NAME"
	BotKeyword           = "BOT_KEYWORD"
	RolePrefix           = "ROLE_PREFIX"
	RoleColor            = "ROLE_COLOR_HEX2DEC"
	Shards               = "SHARDS"
	DiscordrusWebHookURL = "DISCORDRUS_WEBHOOK_URL"
	DiscordBotsOrgBotID  = "DISCORDBOTS_ORG_BOT_ID"
	DiscordBotsOrgToken  = "DISCORDBOTS_ORG_TOKEN" //nolint:gosec // Not a hard-coded credential
)

const (
	defaultLogLevel            = "info"
	defaultLogTimezoneLocation = "UTC"
	defaultPort                = "8080"
	defaultBotName             = "Ephemeral Roles"
	defaultBotKeyword          = "!eph"
	defaultRolePrefix          = "{eph}"
	defaultRoleColor           = "16753920"

	undefinedVariable   = "%s not defined in environment variables"
	integrationDisabled = "integration with discordbots.org disabled"
)

// Variables are variables from the environment.
type Variables struct {
	// Required variables
	BotToken string

	// Optional variables
	LogLevel             string
	LogTimezoneLocation  string
	Port                 string
	BotName              string
	BotKeyword           string
	RolePrefix           string
	RoleColor            string
	DiscordrusWebHookURL string
	DiscordBotsOrgBotID  string
	DiscordBotsOrgToken  string
}

// Lookup looks up expected environment variables and returns a struct
// containing those values.
func Lookup() (*Variables, error) {
	requiredVariables := []string{BotToken}
	requiredVariableValues := make(map[string]string)

	for _, requiredVariable := range requiredVariables {
		value, err := lookupRequired(requiredVariable)
		if err != nil {
			return nil, err
		}

		requiredVariableValues[requiredVariable] = value
	}

	return &Variables{
		BotToken:             requiredVariableValues[BotToken],
		LogLevel:             lookupOptional(LogLevel, defaultLogLevel),
		LogTimezoneLocation:  lookupOptional(LogTimezoneLocation, defaultLogTimezoneLocation),
		Port:                 lookupOptional(Port, defaultPort),
		BotName:              lookupOptional(BotName, defaultBotName),
		BotKeyword:           lookupOptional(BotKeyword, defaultBotKeyword),
		RolePrefix:           lookupOptional(RolePrefix, defaultRolePrefix),
		RoleColor:            lookupOptional(RoleColor, defaultRoleColor),
		DiscordrusWebHookURL: lookupOptional(DiscordrusWebHookURL, ""),
		DiscordBotsOrgBotID:  lookupOptional(DiscordBotsOrgBotID, ""),
		DiscordBotsOrgToken:  lookupOptional(DiscordBotsOrgToken, ""),
	}, nil
}

func lookupRequired(name string) (string, error) {
	value, found := os.LookupEnv(name)
	if !found || value == "" {
		return "", fmt.Errorf(undefinedVariable, name)
	}

	return value, nil
}

func lookupOptional(name, defaultValue string) string {
	value, found := os.LookupEnv(name)
	if !found || value == "" {
		return defaultValue
	}

	return value
}
