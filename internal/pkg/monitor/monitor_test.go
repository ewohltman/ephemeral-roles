package monitor_test

import (
	"context"
	"testing"
	"time"

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
	if err != nil {
		t.Fatal(err)
	}

	metrics := monitor.NewMetrics(&monitor.Config{
		Log:      mock.NewLogger(),
		Session:  session,
		Interval: testMonitorInterval,
	})

	if metrics == nil {
		t.Fatal("Unexpected nil metrics")
	}

	if metrics.ReadyCounter == nil {
		t.Error("Unexpected nil Ready counter")
	}

	if metrics.VoiceStateUpdateCounter == nil {
		t.Error("Unexpected nil VoiceStateUpdate counter")
	}
}

func TestMonitor(t *testing.T) {
	t.Parallel()

	session, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancelCtx := context.WithTimeout(context.Background(), testTimeout)
	defer cancelCtx()

	monitor.NewMetrics(&monitor.Config{
		Log:      mock.NewLogger(),
		Session:  session,
		Interval: testMonitorInterval,
	}).Monitor(ctx)
}
