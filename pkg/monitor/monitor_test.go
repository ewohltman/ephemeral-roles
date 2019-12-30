package monitor

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/mock"
)

const monitorTestInterval = 1 * time.Second

func TestStart(t *testing.T) {
	log := logging.New()
	log.SetOutput(ioutil.Discard)

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
			Interval:            monitorTestInterval,
		},
	)
}
