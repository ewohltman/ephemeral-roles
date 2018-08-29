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
	numMembers float64
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

func Monitor(dgBotSession *discordgo.Session) {
	for {
		check(dgBotSession)
		time.Sleep(time.Second * 5)
	}
}

func check(dgBotSession *discordgo.Session) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	checkNum := cache.numMembers
	numMembers := float64(0)

	for _, guild := range dgBotSession.State.Guilds {
		numMembers += float64(guild.MemberCount)
	}

	if numMembers == checkNum {
		return
	}

	update(numMembers)
}

func update(numMembers float64) {
	cache.numMembers = numMembers
	prometheusMembersGauge.Set(cache.numMembers)
}
