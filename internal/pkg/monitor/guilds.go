package monitor

import (
	"net/http"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/discordbotsorg"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

type guilds struct {
	Log                 logging.Interface
	Session             *discordgo.Session
	HTTPClient          *http.Client
	DiscordBotsOrgBotID string
	DiscordBotsOrgToken string
	PrometheusGauge     prometheus.Gauge
	Interval            time.Duration
	cache               *guildsCache
}

type guildsCache struct {
	mutex     sync.Mutex
	guildList []*discordgo.Guild
	numGuilds int
}

// Monitor sets up an infinite loop checking guild changes
func (g *guilds) Monitor() {
	g.cache = &guildsCache{}

	for {
		g.update()
		time.Sleep(g.Interval)
	}
}

func (g *guilds) update() {
	g.cache.mutex.Lock()
	defer g.cache.mutex.Unlock()

	botName := g.Session.State.User.Username

	originalCount := g.cache.numGuilds
	newCount := len(g.Session.State.Guilds)

	switch {
	case newCount == originalCount:
		return
	case newCount > originalCount:
		newGuild := g.Session.State.Guilds[newCount-1]
		g.Log.WithField("guild", newGuild.Name).Info(botName + " joined new guild")
	case newCount < originalCount:
		g.Log.Info(botName + " removed from guild")
	}

	g.cache.numGuilds = newCount
	g.cache.guildList = g.Session.State.Guilds
	g.PrometheusGauge.Set(float64(newCount))

	// discordbots.org integration
	if g.DiscordBotsOrgBotID != "" && g.DiscordBotsOrgToken != "" {
		err := discordbotsorg.Update(
			g.HTTPClient,
			g.DiscordBotsOrgToken,
			g.DiscordBotsOrgBotID,
			newCount,
		)
		if err != nil {
			g.Log.WithError(err).Warnf("unable to update discordbots.org guild count")
			return
		}
	}
}
