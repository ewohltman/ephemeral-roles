package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

const (
	testMonitorInterval = time.Millisecond
	sleepInterval       = 10 * testMonitorInterval
	testTimeout         = 100 * testMonitorInterval
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
}
