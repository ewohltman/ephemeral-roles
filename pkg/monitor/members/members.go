package members

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
)

type membersCache struct {
	mu         sync.RWMutex
	numMembers int
}

var (
	cache = &membersCache{}

	prometheusMembersGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ephemeral_roles",
			Name:      "count_members",
			Help:      "Total members count",
		},
	)
)

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
