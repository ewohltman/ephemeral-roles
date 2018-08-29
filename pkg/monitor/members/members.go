package members

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
)

type membersCache struct {
	mu         sync.RWMutex
	numMembers int
}

var (
	cache = &membersCache{}
	log   = logging.Instance()

	prometheusMembersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ephemeral_roles",
			Name:      "members_count",
			Help:      "Total Members count",
		},
	)
)

func init() {
	err := prometheus.Register(prometheusMembersGauge)
	if err != nil {
		log.WithError(err).Error("Unable to register Members gauge with Prometheus")
	}
}

func Monitor(dgBotSession *discordgo.Session, token string, botID string) {
	for {
		check(dgBotSession, token, botID)
		time.Sleep(time.Second * 5)
	}
}

func check(dgBotSession *discordgo.Session, token string, botID string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	checkNum := cache.numMembers
	numMembers := 0

	for _, guild := range dgBotSession.State.Guilds {
		numMembers += guild.MemberCount
	}

	if numMembers == checkNum {
		return
	}

	update(numMembers, dgBotSession, token, botID)
}

func update(numMembers int, dgBotSession *discordgo.Session, token string, botID string) {
	cache.numMembers = numMembers

	prometheusMembersGauge.Set(float64(cache.numMembers))
}
