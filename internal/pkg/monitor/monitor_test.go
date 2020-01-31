package monitor

import (
	"testing"
	"time"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const monitorTestInterval = 1 * time.Second

func TestStart(t *testing.T) {
	log := mock.NewLogger()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	Start(
		&Config{
			Log:                 log,
			Session:             session,
			DiscordBotsOrgBotID: "",
			DiscordBotsOrgToken: "",
			Interval:            monitorTestInterval,
		},
	)
}
