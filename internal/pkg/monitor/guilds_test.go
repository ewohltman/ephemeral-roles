package monitor_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

const extraGuildID snowflake.ID = 3000

func TestGuilds_Monitor(t *testing.T) {
	t.Parallel()

	mockSession, err := mock.NewSession()
	require.NoError(t, err)

	log := mock.NewLogger()

	guilds := &monitor.Guilds{
		Log:             log,
		Client:          mockSession,
		Interval:        testMonitorInterval,
		PrometheusGauge: monitor.GuildsGauge(&monitor.Config{Log: log}),
		Cache:           &monitor.GuildsCache{Mutex: &sync.Mutex{}},
	}

	ctx, cancelCtx := context.WithTimeout(t.Context(), testTimeout)
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

	guilds.Client.Caches.AddGuild(discord.Guild{ID: extraGuildID, Name: "testGuildExtra"})
}

func removeGuild(guilds *monitor.Guilds) {
	guilds.Cache.Mutex.Lock()
	defer guilds.Cache.Mutex.Unlock()

	guilds.Client.Caches.RemoveGuild(extraGuildID)
}
