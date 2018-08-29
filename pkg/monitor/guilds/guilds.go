package guilds

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
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
	cache = &guildsCache{}
	log   = logging.Instance()

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

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	buf := bytes.NewBuffer([]byte{})
	for _, guild := range cache.guildList {
		buf.Write([]byte(guild.Name + "\n"))
	}

	response := buf.Bytes()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))

	_, err := w.Write(response)
	if err != nil {
		log.WithError(err).Errorf("Error writing /check HTTP response")
		return
	}
}

func Monitor(dgBotSession *discordgo.Session, token string, botID string) {
	update(dgBotSession, token, botID)

	for {
		check(dgBotSession, token, botID)
		time.Sleep(time.Second * 5)
	}
}

func check(dgBotSession *discordgo.Session, token string, botID string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	checkNum := cache.numGuilds
	numGuilds := len(dgBotSession.State.Guilds)

	if numGuilds == checkNum {
		return
	}

	writeLog(numGuilds, checkNum, dgBotSession)
	update(dgBotSession, token, botID)
}

func writeLog(numGuilds int, checkNum int, dgBotSession *discordgo.Session) {
	if numGuilds > checkNum {
		log.WithField(
			"guild",
			dgBotSession.State.Guilds[len(dgBotSession.State.Guilds)-1].Name,
		).Infof(dgBotSession.State.User.Username + " joined new guild")
	} else {
		log.Infof(dgBotSession.State.User.Username + " removed from guild")
	}
}

func update(dgBotSession *discordgo.Session, token string, botID string) {
	cache.guildList = dgBotSession.State.Guilds
	cache.numGuilds = len(cache.guildList)

	prometheusGuildsGauge.Set(float64(cache.numGuilds))

	// discordbots.org integration
	if token != "" && botID != "" {
		response, err := discordBotsOrg.Update(token, botID, cache.numGuilds)
		if err != nil {
			log.WithError(err).Warnf("unable to update guild count")
			return
		}

		if response != "{}" {
			log.WithField("response", response).Warnf("discordbots.org integration: abnormal response")
		}
	}
}
