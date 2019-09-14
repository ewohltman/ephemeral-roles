package monitor

import (
	"testing"
	"time"

	"github.com/ewohltman/ephemeral-roles/pkg/mock"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

func TestStart(t *testing.T) {
	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

	mockSession, err := mock.Session()
	if err != nil {
		t.Fatal(err)
	}

	Start(
		&Config{
			Log:                 log,
			Session:             mockSession,
			BotID:               "",
			DiscordBotsOrgToken: "",
			Interval:            1 * time.Second,
		},
	)
}
