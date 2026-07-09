// Package monitor provides implementations for monitoring statistics and
// exposing them as Prometheus metrics.
package monitor

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/prometheus/client_golang/prometheus"
)

const prometheusNamespace = "ephemeral_roles"

// Config contains fields for configuring Metrics.
type Config struct {
	Log      *slog.Logger
	Client   *bot.Client
	Interval time.Duration
}

// Metrics contains fields for tracking and exposing metrics to Prometheus.
type Metrics struct {
	*Config

	Guilds                  *Guilds
	Members                 *Members
	ReadyCounter            prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
	GuildsGauge             prometheus.Gauge
	MembersGauge            prometheus.Gauge
}

// NewMetrics returns a new *Metrics configured using the provided config.
func NewMetrics(config *Config) *Metrics {
	metrics := &Metrics{
		Config:                  config,
		ReadyCounter:            ReadyCounter(config),
		VoiceStateUpdateCounter: VoiceStateUpdateCounter(config),
		GuildsGauge:             GuildsGauge(config),
		MembersGauge:            MembersGauge(config),
	}

	metrics.newGuilds()
	metrics.newMembers()

	return metrics
}

// Monitor begins the goroutines for monitoring callback metrics.
func (metrics *Metrics) Monitor(ctx context.Context) {
	go metrics.Guilds.Monitor(ctx)
	go metrics.Members.Monitor(ctx)
}

func (metrics *Metrics) newGuilds() {
	metrics.Guilds = &Guilds{
		Log:             metrics.Log,
		Client:          metrics.Client,
		Interval:        metrics.Interval,
		PrometheusGauge: metrics.GuildsGauge,
		Cache:           &GuildsCache{Mutex: &sync.Mutex{}},
	}
}

func (metrics *Metrics) newMembers() {
	metrics.Members = &Members{
		Log:             metrics.Log,
		Client:          metrics.Client,
		Interval:        metrics.Interval,
		PrometheusGauge: metrics.MembersGauge,
		Cache:           &MembersCache{Mutex: &sync.Mutex{}},
	}
}

// ReadyCounter returns a Prometheus counter for Ready events.
func ReadyCounter(config *Config) prometheus.Counter {
	prometheusReadyCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Name:      "ready_events_total",
			Help:      "Total Ready events",
		},
	)

	if err := prometheus.Register(prometheusReadyCounter); err != nil && !alreadyRegisteredError(err) {
		config.Log.Error("Unable to register Ready events metric with Prometheus", "error", err)
		return nil
	}

	return prometheusReadyCounter
}

// VoiceStateUpdateCounter returns a Prometheus counter for VoiceStateUpdate
// events.
func VoiceStateUpdateCounter(config *Config) prometheus.Counter {
	prometheusVoiceStateUpdateCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Name:      "voice_state_update_events_total",
			Help:      "Total VoiceStateUpdate events",
		},
	)

	if err := prometheus.Register(prometheusVoiceStateUpdateCounter); err != nil && !alreadyRegisteredError(err) {
		config.Log.Error("Unable to register VoiceStateUpdate events metric with Prometheus", "error", err)
		return nil
	}

	return prometheusVoiceStateUpdateCounter
}

// GuildsGauge returns a Prometheus gauge for the number of guilds the bot
// belongs to.
func GuildsGauge(config *Config) prometheus.Gauge {
	prometheusGuildsGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Name:      "guilds",
			Help:      "Total Guilds count",
		},
	)

	if err := prometheus.Register(prometheusGuildsGauge); err != nil && !alreadyRegisteredError(err) {
		config.Log.Error("Unable to register Guilds gauge with Prometheus", "error", err)
		return nil
	}

	return prometheusGuildsGauge
}

// MembersGauge returns a Prometheus gauge for the number of members of the
// guilds the bot belongs to.
func MembersGauge(config *Config) prometheus.Gauge {
	prometheusMembersGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: prometheusNamespace,
			Name:      "members",
			Help:      "Total Members count",
		},
	)

	if err := prometheus.Register(prometheusMembersGauge); err != nil && !alreadyRegisteredError(err) {
		config.Log.Error("Unable to register Members gauge with Prometheus", "error", err)
		return nil
	}

	return prometheusMembersGauge
}

func alreadyRegisteredError(err error) bool {
	return errors.As(err, &prometheus.AlreadyRegisteredError{})
}
