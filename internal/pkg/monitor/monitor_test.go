package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

const (
	testMonitorInterval = time.Millisecond
	sleepDuration       = 5 * testMonitorInterval
)

func TestMetrics(t *testing.T) {
	config, mockSession, err := newTestConfig()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, mockSession)

	metrics := Metrics(config)

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

func TestStart(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	config, mockSession, err := newTestConfig()
	if err != nil {
		t.Fatal(err)
	}

	defer mock.SessionClose(t, mockSession)

	Start(ctx, config)
}

func newTestConfig() (*Config, *discordgo.Session, error) {
	mockSession, err := mock.NewSession()
	if err != nil {
		return nil, nil, err
	}

	config := &Config{
		Log:      mock.NewLogger(),
		Session:  mockSession,
		Interval: testMonitorInterval,
	}

	return config, mockSession, nil
}
