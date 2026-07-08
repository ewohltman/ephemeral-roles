package monitor

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

// Guilds contains fields for monitoring the guilds the bot belongs to.
type Guilds struct {
	Log             *slog.Logger
	Session         *discordgo.Session
	Interval        time.Duration
	PrometheusGauge prometheus.Gauge
	Cache           *GuildsCache
}

// GuildsCache is an in-memory cache of the guilds the bot belongs to.
type GuildsCache struct {
	Mutex     *sync.Mutex
	guildList []*discordgo.Guild
	numGuilds int
}

// Monitor sets up an infinite loop checking guild changes.
func (guilds *Guilds) Monitor(ctx context.Context) {
	updateTicker := time.NewTicker(guilds.Interval)
	defer updateTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-updateTicker.C:
			guilds.update()
		}
	}
}

func (guilds *Guilds) update() {
	guilds.Cache.Mutex.Lock()
	defer guilds.Cache.Mutex.Unlock()

	originalCount := guilds.Cache.numGuilds
	newCount := len(guilds.Session.State.Guilds)

	switch {
	case newCount == originalCount:
		return
	case newCount > originalCount && originalCount != 0:
		botName := guilds.Session.State.User.Username
		newGuild := guilds.Session.State.Guilds[newCount-1]
		guilds.Log.Info(botName+" joined new guild", "guild", newGuild.Name)
	case newCount < originalCount:
		botName := guilds.Session.State.User.Username
		guilds.Log.Info(botName + " removed from guild")
	}

	guilds.Cache.numGuilds = newCount
	guilds.Cache.guildList = guilds.Session.State.Guilds
	guilds.PrometheusGauge.Set(float64(newCount))
}
