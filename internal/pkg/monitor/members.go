package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

type members struct {
	Log                 logging.Interface
	Session             *discordgo.Session
	DiscordBotsOrgBotID string
	DiscordBotsOrgToken string
	PrometheusGauge     prometheus.Gauge
	Interval            time.Duration
	cache               *membersCache
}

type membersCache struct {
	mutex      sync.Mutex
	numMembers int
}

// Monitor sets up an infinite loop checking member changes
func (m *members) Monitor(ctx context.Context) (done chan struct{}) {
	done = make(chan struct{})
	defer close(done)

	m.cache = &membersCache{}

	updateTicker := time.NewTicker(m.Interval)
	defer updateTicker.Stop()

	for {
		select {
		case <-updateTicker.C:
			m.update()
		case <-ctx.Done():
			return
		}
	}
}

func (m *members) update() {
	m.cache.mutex.Lock()
	defer m.cache.mutex.Unlock()

	numMembers := 0

	for _, guild := range m.Session.State.Guilds {
		numMembers += guild.MemberCount
	}

	if numMembers != m.cache.numMembers {
		m.cache.numMembers = numMembers
		m.PrometheusGauge.Set(float64(m.cache.numMembers))
	}
}
