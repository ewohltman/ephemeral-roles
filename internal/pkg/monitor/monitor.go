// Package monitor provides implementations for monitoring statistics and
// exposing them as Prometheus metrics.
package monitor

import (
	"context"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

// Config contains fields for the monitoring methods.
type Config struct {
	Log                 logging.Interface
	Session             *discordgo.Session
	HTTPClient          *http.Client
	DiscordBotsOrgBotID string
	DiscordBotsOrgToken string
	Interval            time.Duration
}

// Metrics returns back configured *CallbackMetrics.
func Metrics(config *Config) *CallbackMetrics {
	return &CallbackMetrics{
		ReadyCounter:            config.ReadyCounter(),
		MessageCreateCounter:    config.MessageCreateCounter(),
		VoiceStateUpdateCounter: config.VoiceStateUpdateCounter(),
	}
}

// Start begins the goroutines for monitoring metrics.
func Start(ctx context.Context, config *Config) {
	go config.guilds().Monitor(ctx)

	go config.members().Monitor(ctx)
}

// CallbackMetrics are Prometheus objects for tracking and exposing metrics.
type CallbackMetrics struct {
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}

// ReadyCounter returns a prometheus.Counter for Discord API Ready events.
func (config *Config) ReadyCounter() prometheus.Counter {
	prometheusReadyCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ephemeral_roles",
			Name:      "ready_events",
			Help:      "Total Ready events",
		},
	)

	err := prometheus.Register(prometheusReadyCounter)
	if err != nil {
		config.Log.WithError(err).Error("Unable to register Ready events metric with Prometheus")
		return nil
	}

	return prometheusReadyCounter
}

// MessageCreateCounter returns a prometheus.Counter for Discord API
// MessageCreate events.
func (config *Config) MessageCreateCounter() prometheus.Counter {
	prometheusMessageCreateCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ephemeral_roles",
			Name:      "message_create_events",
			Help:      "Total MessageCreate events",
		},
	)

	err := prometheus.Register(prometheusMessageCreateCounter)
	if err != nil {
		config.Log.WithError(err).Error("Unable to register MessageCreate events metric with Prometheus")
		return nil
	}

	return prometheusMessageCreateCounter
}

// VoiceStateUpdateCounter returns a prometheus.Counter for Discord API
// VoiceStateUpdate events.
func (config *Config) VoiceStateUpdateCounter() prometheus.Counter {
	prometheusVoiceStateUpdateCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "ephemeral_roles",
			Name:      "voice_state_update_events",
			Help:      "Total VoiceStateUpdate events",
		},
	)

	err := prometheus.Register(prometheusVoiceStateUpdateCounter)
	if err != nil {
		config.Log.WithError(err).Error("Unable to register VoiceStateUpdate events metric with Prometheus")
		return nil
	}

	return prometheusVoiceStateUpdateCounter
}

func (config *Config) guilds() *guilds {
	prometheusGuildsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ephemeral_roles",
			Name:      "guilds_count",
			Help:      "Total Guilds count",
		},
	)

	err := prometheus.Register(prometheusGuildsGauge)
	if err != nil {
		config.Log.WithError(err).Error("Unable to register Guilds gauge with Prometheus")
	}

	return &guilds{
		Log:                 config.Log,
		HTTPClient:          config.HTTPClient,
		Session:             config.Session,
		DiscordBotsOrgBotID: config.DiscordBotsOrgBotID,
		DiscordBotsOrgToken: config.DiscordBotsOrgToken,
		PrometheusGauge:     prometheusGuildsGauge,
		Interval:            config.Interval,
	}
}

func (config *Config) members() *members {
	prometheusMembersGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "ephemeral_roles",
			Name:      "members_count",
			Help:      "Total Members count",
		},
	)

	err := prometheus.Register(prometheusMembersGauge)
	if err != nil {
		config.Log.WithError(err).Error("Unable to register Members gauge with Prometheus")
	}

	return &members{
		Log:                 config.Log,
		Session:             config.Session,
		DiscordBotsOrgBotID: config.DiscordBotsOrgBotID,
		DiscordBotsOrgToken: config.DiscordBotsOrgToken,
		PrometheusGauge:     prometheusMembersGauge,
		Interval:            config.Interval,
	}
}
