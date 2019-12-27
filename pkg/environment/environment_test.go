package environment

import (
	"os"
	"testing"
)

func TestCheckRequiredVariables(t *testing.T) {
	envVars := []string{
		Port,
		BotToken,
		BotName,
		BotKeyword,
		RolePrefix,
		RoleColor,
	}

	setTestEnvironmentVariables(t, envVars)
	defer unSetTestEnvironmentVariables(t, envVars)

	_, err := CheckRequiredVariables()
	if err != nil {
		t.Errorf("Error checking required environment variables: %s", err)
	}
}

func TestCheckOptionalVariables(t *testing.T) {
	envVars := []string{
		LogLevel,
		DiscordrusWebHookURL,
		LogTimezoneLocation,
		DiscordBotsOrgBotID,
		DiscordBotsOrgToken,
	}

	setTestEnvironmentVariables(t, envVars)
	defer unSetTestEnvironmentVariables(t, envVars)

	_, err := CheckOptionalVariables()
	if err != nil {
		t.Errorf("Error checking optional environment variables: %s", err)
	}
}

func setTestEnvironmentVariables(t *testing.T, envVars []string) {
	for _, envVar := range envVars {
		err := os.Setenv(envVar, "testVal")
		if err != nil {
			t.Errorf("Error setting test environment variable: %s", err)
		}
	}
}

func unSetTestEnvironmentVariables(t *testing.T, envVars []string) {
	for _, envVar := range envVars {
		err := os.Unsetenv(envVar)
		if err != nil {
			t.Errorf("Error unsetting test environment variable: %s", err)
		}
	}
}
