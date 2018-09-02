// Package callbacks is a collection of the callback functions used in response
// to events from Discord's Websocket (WS)API.  Common definitions across the
// package are contained within common.go
package callbacks

import (
	"os"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// BOTNAME is the name of the bot
	BOTNAME = os.Getenv("BOT_NAME")

	// BOTKEYWORD is the message prefix the bot should watch for
	BOTKEYWORD = os.Getenv("BOT_KEYWORD") + " "

	// ROLEPREFIX is the prefix to add before ephemeral role names
	ROLEPREFIX = os.Getenv("ROLE_PREFIX") + " "

	log = logging.Instance()
)

func init() {
	err := prometheus.Register(prometheusReadyCounter)
	if err != nil {
		log.WithError(err).Error("Unable to register Ready events metric with Prometheus")
	}

	err = prometheus.Register(prometheusVoiceStateUpdateCounter)
	if err != nil {
		log.WithError(err).Error("Unable to register VoiceStateUpdate events metric with Prometheus")
	}

	err = prometheus.Register(prometheusMessageCreateCounter)
	if err != nil {
		log.WithError(err).Error("Unable to register MessageCreate events metric with Prometheus")
	}
}
