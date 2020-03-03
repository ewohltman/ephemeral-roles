// Package environment provides implementations for parsing expected
// environment variables.
package environment

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	DiscordrusWebHookURL = "DISCORDRUS_WEBHOOK_URL"
	DiscordBotsOrgBotID  = "DISCORDBOTS_ORG_BOT_ID"
	DiscordBotsOrgToken  = "DISCORDBOTS_ORG_TOKEN" //nolint:gosec // Not a hard-coded credential
	InstanceName         = "INSTANCE_NAME"
	ShardCount           = "SHARD_COUNT"
)

const (
	defaultLogLevel            = "info"
	defaultLogTimezoneLocation = "UTC"
	defaultPort                = "8080"
	defaultBotName             = "Ephemeral Roles"
	defaultBotKeyword          = "!eph"
	defaultRolePrefix          = "{eph}"
	defaultRoleColor           = "16753920"
	defaultInstanceName        = "ephemeral-roles-0"
	defaultShardCount          = "1"

	undefinedVariable = "%s not defined in environment variables"
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
	RoleColor            int
	DiscordrusWebHookURL string
	DiscordBotsOrgBotID  string
	DiscordBotsOrgToken  string
	InstanceName         string
	ShardID              int
	ShardCount           int
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

	shardID, err := parseShardID()
	if err != nil {
		return nil, err
	}

	roleColor, err := lookupOptionalInt(RoleColor, defaultRoleColor)
	if err != nil {
		return nil, err
	}

	shardCount, err := lookupOptionalInt(ShardCount, defaultShardCount)
	if err != nil {
		return nil, err
	}

	return &Variables{
		BotToken:             requiredVariableValues[BotToken],
		LogLevel:             lookupOptional(LogLevel, defaultLogLevel),
		LogTimezoneLocation:  lookupOptional(LogTimezoneLocation, defaultLogTimezoneLocation),
		Port:                 lookupOptional(Port, defaultPort),
		BotName:              lookupOptional(BotName, defaultBotName),
		BotKeyword:           lookupOptional(BotKeyword, defaultBotKeyword),
		RolePrefix:           lookupOptional(RolePrefix, defaultRolePrefix),
		RoleColor:            roleColor,
		DiscordrusWebHookURL: lookupOptional(DiscordrusWebHookURL, ""),
		DiscordBotsOrgBotID:  lookupOptional(DiscordBotsOrgBotID, ""),
		DiscordBotsOrgToken:  lookupOptional(DiscordBotsOrgToken, ""),
		InstanceName:         lookupOptional(InstanceName, defaultInstanceName),
		ShardID:              shardID,
		ShardCount:           shardCount,
	}, nil
}

func lookupRequired(name string) (string, error) {
	value, found := os.LookupEnv(name)
	if !found || value == "" {
		return "", fmt.Errorf(undefinedVariable, name)
	}

	return value, nil
}

func lookupOptionalInt(name, defaultValue string) (int, error) {
	value := lookupOptional(name, defaultValue)

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"error converting %s (%s) to int: %s",
			RoleColor,
			value,
			err,
		)
	}

	return intValue, nil
}

func lookupOptional(name, defaultValue string) string {
	value, found := os.LookupEnv(name)
	if !found || value == "" {
		return defaultValue
	}

	return value
}

func parseShardID() (int, error) {
	instanceName := lookupOptional(InstanceName, defaultInstanceName)

	shardIDRegEx := regexp.MustCompile(`-[0-9].*$`)

	shardIDString := shardIDRegEx.FindString(instanceName)
	shardIDString = strings.TrimPrefix(shardIDString, "-")

	shardID, err := strconv.Atoi(shardIDString)
	if err != nil {
		return 0, fmt.Errorf(
			"error parsing shard ID: %s",
			err,
		)
	}

	return shardID, nil
}
