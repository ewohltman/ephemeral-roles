package config

import "testing"

func TestCheckRequired(t *testing.T) {
	_, _, err := CheckRequired()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckOptional(t *testing.T) {
	_, _, err := CheckDiscordBotsOrg()
	if err != nil {
		t.Log(err)
	}
}
