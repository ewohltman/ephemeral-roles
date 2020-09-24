package monitor_test

import (
	"context"
	"sync"
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

func TestMembers_Monitor(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), testTimeout)
	defer cancelCtx()

	mockSession, err := mock.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, mockSession)

	log := mock.NewLogger()

	members := &monitor.Members{
		Log:             log,
		Session:         mockSession,
		Interval:        testMonitorInterval,
		PrometheusGauge: monitor.MembersGauge(&monitor.Config{Log: log}),
		Cache:           &monitor.MembersCache{Mutex: &sync.Mutex{}},
	}

	go func() {
		members.Monitor(ctx)
	}()

	<-ctx.Done()
}
