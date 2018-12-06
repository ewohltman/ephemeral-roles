package guilds

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/config"
	"github.com/ewohltman/ephemeral-roles/pkg/discordBotsOrg"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
)

type guildsCache struct {
	mu        sync.RWMutex
	guildList []*discordgo.Guild
	numGuilds int
}

var (
	botID               = ""
	discordBotsOrgToken = ""
	cache               = &guildsCache{}
	log                 = logging.Instance()

	prometheusGuildsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ephemeral_roles",
			Name:      "guilds_count",
			Help:      "Total Guilds count",
		},
	)
)

func init() {
	err := prometheus.Register(prometheusGuildsGauge)
	if err != nil {
		log.WithError(err).Error("Unable to register Guilds gauge with Prometheus")
	}
}

// Monitor sets up an infinite loop checking guild changes
func Monitor(dgBotSession *discordgo.Session) {
	var err error

	botID, discordBotsOrgToken, err = config.CheckDiscordBotsOrg()
	if err != nil {
		botID = ""
		discordBotsOrgToken = ""
	}

	update(dgBotSession)

	for {
		check(dgBotSession)
		time.Sleep(time.Second * 5)
	}
}

func update(dgBotSession *discordgo.Session) {
	cache.guildList = dgBotSession.State.Guilds
	cache.numGuilds = len(cache.guildList)
	prometheusGuildsGauge.Set(float64(cache.numGuilds))

	// discordbots.org integration
	if botID != "" && discordBotsOrgToken != "" {
		err := discordBotsOrg.Update(discordBotsOrgToken, botID, cache.numGuilds)
		if err != nil {
			log.WithError(err).Warnf("unable to update guild count")
			return
		}
	}
}

func check(dgBotSession *discordgo.Session) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	numGuilds := len(dgBotSession.State.Guilds)

	if numGuilds == cache.numGuilds {
		return
	}

	writeLog(numGuilds, dgBotSession)
	update(dgBotSession)
}

func writeLog(numGuilds int, dgBotSession *discordgo.Session) {
	if numGuilds > cache.numGuilds {
		log.WithField(
			"guild",
			dgBotSession.State.Guilds[len(dgBotSession.State.Guilds)-1].Name,
		).Infof(dgBotSession.State.User.Username + " joined new guild")
	} else {
		log.Infof(dgBotSession.State.User.Username + " removed from guild")
	}
}
