package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestGuilds_Monitor(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())

	config, mockSession, err := newTestConfig()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, mockSession)

	guilds := config.guilds()

	go func() {
		time.Sleep(sleepDuration)

		cancelCtx()
	}()

	<-guilds.Monitor(ctx)
}
