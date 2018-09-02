package callbacks

import "testing"

func TestDiscordError_Error(t *testing.T) {
	log.WithError(&discordError{
		CustomMessage: "custom test message",
	}).Info("test discord error")
}
