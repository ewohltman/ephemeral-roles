package monitor

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type members struct {
	Log                 *logrus.Logger
	Session             *discordgo.Session
	BotID               string
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
func (m *members) Monitor() {
	m.cache = &membersCache{}

	for {
		m.update()
		time.Sleep(m.Interval)
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
