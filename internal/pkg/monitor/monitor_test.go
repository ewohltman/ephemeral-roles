package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

const (
	testMonitorInterval = time.Millisecond
	sleepInterval       = 10 * testMonitorInterval
	testTimeout         = 100 * testMonitorInterval

	newGuildID snowflake.ID = 987654321
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	metrics := monitor.NewMetrics(&monitor.Config{
		Log:      mock.NewLogger(),
		Client:   session,
		Interval: testMonitorInterval,
	})

	require.NotNil(t, metrics)
	assert.NotNil(t, metrics.ReadyCounter)
	assert.NotNil(t, metrics.VoiceStateUpdateCounter)
	assert.NotNil(t, metrics.GuildsGauge)
	assert.NotNil(t, metrics.MembersGauge)
}

func TestMonitor(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	require.NoError(t, err)

	ctx, cancelCtx := context.WithTimeout(t.Context(), testTimeout)
	defer cancelCtx()

	monitor.NewMetrics(&monitor.Config{
		Log:      mock.NewLogger(),
		Client:   session,
		Interval: testMonitorInterval,
	}).Monitor(ctx)

	// Drive the guild-join and guild-removal branches: the monitor
	// goroutine observes the cache growing and shrinking between ticks.
	time.Sleep(sleepInterval)

	session.Caches.AddGuild(discord.Guild{ID: newGuildID, Name: "newGuild", MemberCount: 5})

	time.Sleep(sleepInterval)

	session.Caches.RemoveGuild(newGuildID)

	<-ctx.Done()
}
