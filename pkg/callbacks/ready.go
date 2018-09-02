package callbacks

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var prometheusReadyCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Namespace: "ephemeral_roles",
		Name:      "ready_events",
		Help:      "Total Ready events",
	},
)

// Ready is the callback function for the Ready event from Discord
func Ready(s *discordgo.Session, event *discordgo.Ready) {
	// Increment the total number of Ready events
	prometheusReadyCounter.Inc()

	log.WithFields(logrus.Fields{
		"server_count": len(event.Guilds),
	}).Infof("\"" + BOTNAME + "\" Ready")

	idleSince := 0

	usd := discordgo.UpdateStatusData{
		IdleSince: &idleSince,
		Game: &discordgo.Game{
			Name: strings.TrimSpace(BOTKEYWORD),
			Type: discordgo.GameTypeWatching,
		},
		AFK:    false,
		Status: "online",
	}

	err := s.UpdateStatusComplex(usd)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"UpdateStatusData": usd,
		}).Errorf("Error updating complex status")
	}
}
