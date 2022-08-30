package monitor_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestGuilds_Monitor(t *testing.T) {
	t.Parallel()

	mockSession, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	log := mock.NewLogger()

	guilds := &monitor.Guilds{
		Log:             log,
		Session:         mockSession,
		Interval:        testMonitorInterval,
		PrometheusGauge: monitor.GuildsGauge(&monitor.Config{Log: log}),
		Cache:           &monitor.GuildsCache{Mutex: &sync.Mutex{}},
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), testTimeout)
	defer cancelCtx()

	go func() {
		guilds.Monitor(ctx)
	}()

	time.Sleep(sleepInterval)

	addGuild(guilds)

	time.Sleep(sleepInterval)

	removeGuild(guilds)

	<-ctx.Done()
}

func addGuild(guilds *monitor.Guilds) {
	guilds.Cache.Mutex.Lock()
	defer guilds.Cache.Mutex.Unlock()

	guilds.Session.State.Guilds = append(guilds.Session.State.Guilds, &discordgo.Guild{})
}

func removeGuild(guilds *monitor.Guilds) {
	guilds.Cache.Mutex.Lock()
	defer guilds.Cache.Mutex.Unlock()

	guilds.Session.State.Guilds = guilds.Session.State.Guilds[:len(guilds.Session.State.Guilds)-1]
}
