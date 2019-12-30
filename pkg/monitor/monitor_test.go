package monitor

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/mock"
)

func TestStart(t *testing.T) {
	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

	session, err := mock.Session()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, session)

	Start(
		&Config{
			Log:                 log,
			Session:             session,
			BotID:               "",
			DiscordBotsOrgToken: "",
			Interval:            1 * time.Second,
		},
	)
}
