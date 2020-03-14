package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

type guilds struct {
	Log             logging.Interface
	Session         *discordgo.Session
	PrometheusGauge prometheus.Gauge
	Interval        time.Duration
	cache           *guildsCache
}

type guildsCache struct {
	mutex     sync.Mutex
	guildList []*discordgo.Guild
	numGuilds int
}

// Monitor sets up an infinite loop checking guild changes
func (g *guilds) Monitor(ctx context.Context) (done chan struct{}) {
	done = make(chan struct{})
	defer close(done)

	g.cache = &guildsCache{}

	updateTicker := time.NewTicker(g.Interval)
	defer updateTicker.Stop()

	for {
		select {
		case <-updateTicker.C:
			g.update()
		case <-ctx.Done():
			return
		}
	}
}

func (g *guilds) update() {
	g.cache.mutex.Lock()
	defer g.cache.mutex.Unlock()

	originalCount := g.cache.numGuilds
	newCount := len(g.Session.State.Guilds)

	switch {
	case newCount == originalCount:
		return
	case newCount > originalCount && originalCount != 0:
		botName := g.Session.State.User.Username
		newGuild := g.Session.State.Guilds[newCount-1]
		g.Log.WithField("guild", newGuild.Name).Info(botName + " joined new guild")
	case newCount < originalCount:
		botName := g.Session.State.User.Username
		g.Log.Info(botName + " removed from guild")
	}

	g.cache.numGuilds = newCount
	g.cache.guildList = g.Session.State.Guilds
	g.PrometheusGauge.Set(float64(newCount))
}
