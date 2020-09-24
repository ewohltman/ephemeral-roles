package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

// Members contains fields for monitoring the number of members in the guilds
// the bot belongs to.
type Members struct {
	Log             logging.Interface
	Session         *discordgo.Session
	Interval        time.Duration
	PrometheusGauge prometheus.Gauge
	Cache           *MembersCache
}

// MembersCache is an in-memory cache of the number of members in the guilds
// the bot belongs to.
type MembersCache struct {
	Mutex      *sync.Mutex
	numMembers int
}

// Monitor sets up an infinite loop checking member changes.
func (members *Members) Monitor(ctx context.Context) {
	updateTicker := time.NewTicker(members.Interval)
	defer updateTicker.Stop()

	for {
		select {
		case <-updateTicker.C:
			members.update()
		case <-ctx.Done():
			return
		}
	}
}

func (members *Members) update() {
	members.Cache.Mutex.Lock()
	defer members.Cache.Mutex.Unlock()

	numMembers := 0

	for _, guild := range members.Session.State.Guilds {
		numMembers += guild.MemberCount
	}

	if numMembers != members.Cache.numMembers {
		members.Cache.numMembers = numMembers
		members.PrometheusGauge.Set(float64(members.Cache.numMembers))
	}
}
