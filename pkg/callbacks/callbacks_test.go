package callbacks

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

func TestDiscordError_Error(t *testing.T) {
	log := logging.New()
	log.SetOutput(ioutil.Discard)

	err := &discordError{
		HTTPResponseMessage: "test HTTP error response message",
		APIResponse: &DiscordAPIResponse{
			Code:    http.StatusInternalServerError,
			Message: "test Discord error response message",
		},
		CustomMessage: "test error message",
	}

	log.WithField("error", err.Error()).Info("test discord error")
}

func TestDiscordError_String(t *testing.T) {
	log := logging.New()
	log.SetOutput(ioutil.Discard)

	err := &discordError{
		HTTPResponseMessage: "test HTTP error response message",
		APIResponse: &DiscordAPIResponse{
			Code:    http.StatusInternalServerError,
			Message: "test Discord error response message",
		},
		CustomMessage: "test error message",
	}

	log.WithField("error", err.String()).Info("test discord error")
}
