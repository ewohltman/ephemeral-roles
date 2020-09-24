package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
)

const (
	testMonitorInterval = time.Millisecond
	sleepInterval       = 10 * testMonitorInterval
	testTimeout         = 100 * testMonitorInterval
)

func TestMetrics(t *testing.T) {
	metrics, mockSession, err := newTestMetrics()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, mockSession)

	if metrics == nil {
		t.Fatal("Unexpected nil metrics")
	}

	if metrics.ReadyCounter == nil {
		t.Error("Unexpected nil Ready counter")
	}

	if metrics.MessageCreateCounter == nil {
		t.Error("Unexpected nil MessageCreate counter")
	}

	if metrics.VoiceStateUpdateCounter == nil {
		t.Error("Unexpected nil VoiceStateUpdate counter")
	}
}

func TestMonitor(t *testing.T) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), testTimeout)
	defer cancelCtx()

	metrics, mockSession, err := newTestMetrics()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, mockSession)

	metrics.Monitor(ctx)
}

func newTestMetrics() (*monitor.Metrics, *discordgo.Session, error) {
	mockSession, err := mock.NewSession()
	if err != nil {
		return nil, nil, err
	}

	config := &monitor.Config{
		// Log:      mock.NewLogger(),
		Log:      logging.New(),
		Session:  mockSession,
		Interval: testMonitorInterval,
	}

	metrics := monitor.NewMetrics(config)

	return metrics, mockSession, nil
}
