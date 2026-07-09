package monitor

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/prometheus/client_golang/prometheus"
)

// Guilds contains fields for monitoring the guilds the bot belongs to.
type Guilds struct {
	Log             *slog.Logger
	Client          *bot.Client
	Interval        time.Duration
	PrometheusGauge prometheus.Gauge
	Cache           *GuildsCache
}

// GuildsCache is an in-memory cache of the guilds the bot belongs to.
type GuildsCache struct {
	Mutex     *sync.Mutex
	guildList []discord.Guild
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

	currentGuilds := slices.Collect(guilds.Client.Caches.Guilds())

	originalCount := guilds.Cache.numGuilds
	newCount := len(currentGuilds)

	switch {
	case newCount == originalCount:
		return
	case newCount > originalCount && originalCount != 0:
		if newGuild, ok := guilds.newlyJoinedGuild(currentGuilds); ok {
			guilds.Log.Info(guilds.botName()+" joined new guild", "guild", newGuild.Name)
		}
	case newCount < originalCount:
		guilds.Log.Info(guilds.botName() + " removed from guild")
	}

	guilds.Cache.numGuilds = newCount
	guilds.Cache.guildList = currentGuilds
	guilds.PrometheusGauge.Set(float64(newCount))
}

func (guilds *Guilds) botName() string {
	selfUser, ok := guilds.Client.Caches.SelfUser()
	if !ok {
		return ""
	}

	return selfUser.Username
}

func (guilds *Guilds) newlyJoinedGuild(currentGuilds []discord.Guild) (discord.Guild, bool) {
	for i := range currentGuilds {
		if !guilds.isKnownGuild(currentGuilds[i].ID) {
			return currentGuilds[i], true
		}
	}

	return discord.Guild{}, false
}

func (guilds *Guilds) isKnownGuild(guildID snowflake.ID) bool {
	for i := range guilds.Cache.guildList {
		if guilds.Cache.guildList[i].ID == guildID {
			return true
		}
	}

	return false
}
