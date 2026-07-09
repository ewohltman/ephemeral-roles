// Package monitor provides implementations for monitoring statistics and
// exposing them as Prometheus metrics.
package monitor

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
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

	ReadyCounter            prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
	GuildsGauge             prometheus.Gauge
	MembersGauge            prometheus.Gauge

	guildList  []discord.Guild
	numGuilds  int
	numMembers int
}

// NewMetrics returns a new *Metrics configured using the provided config.
func NewMetrics(config *Config) *Metrics {
	return &Metrics{
		Config:                  config,
		ReadyCounter:            ReadyCounter(config),
		VoiceStateUpdateCounter: VoiceStateUpdateCounter(config),
		GuildsGauge:             GuildsGauge(config),
		MembersGauge:            MembersGauge(config),
	}
}

// Monitor periodically polls the client cache and updates the guild and member
// gauges until the context is canceled. It blocks, so callers should invoke it
// in its own goroutine (go metrics.Monitor(ctx)).
func (metrics *Metrics) Monitor(ctx context.Context) {
	updateTicker := time.NewTicker(metrics.Interval)
	defer updateTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-updateTicker.C:
			metrics.updateGuilds()
			metrics.updateMembers()
		}
	}
}

func (metrics *Metrics) updateGuilds() {
	newCount := metrics.Client.Caches.GuildsLen()
	if newCount == metrics.numGuilds {
		return
	}

	currentGuilds := slices.Collect(metrics.Client.Caches.Guilds())

	switch {
	case newCount > metrics.numGuilds && metrics.numGuilds != 0:
		if newGuild, ok := metrics.newlyJoinedGuild(currentGuilds); ok {
			metrics.Log.Info(metrics.botName()+" joined new guild", "guild", newGuild.Name)
		}
	case newCount < metrics.numGuilds:
		metrics.Log.Info(metrics.botName() + " removed from guild")
	}

	metrics.numGuilds = newCount
	metrics.guildList = currentGuilds
	metrics.GuildsGauge.Set(float64(newCount))
}

func (metrics *Metrics) updateMembers() {
	numMembers := 0

	for guild := range metrics.Client.Caches.Guilds() {
		numMembers += guild.MemberCount
	}

	if numMembers != metrics.numMembers {
		metrics.numMembers = numMembers
		metrics.MembersGauge.Set(float64(numMembers))
	}
}

func (metrics *Metrics) botName() string {
	selfUser, ok := metrics.Client.Caches.SelfUser()
	if !ok {
		return ""
	}

	return selfUser.Username
}

func (metrics *Metrics) newlyJoinedGuild(currentGuilds []discord.Guild) (discord.Guild, bool) {
	for i := range currentGuilds {
		if !metrics.isKnownGuild(currentGuilds[i].ID) {
			return currentGuilds[i], true
		}
	}

	return discord.Guild{}, false
}

func (metrics *Metrics) isKnownGuild(guildID snowflake.ID) bool {
	for i := range metrics.guildList {
		if metrics.guildList[i].ID == guildID {
			return true
		}
	}

	return false
}

// ReadyCounter returns a Prometheus counter for Ready events.
func ReadyCounter(config *Config) prometheus.Counter {
	return newCounter(config.Log, "ready_events_total", "Total Ready events")
}

// VoiceStateUpdateCounter returns a Prometheus counter for VoiceStateUpdate
// events.
func VoiceStateUpdateCounter(config *Config) prometheus.Counter {
	return newCounter(config.Log, "voice_state_update_events_total", "Total VoiceStateUpdate events")
}

// GuildsGauge returns a Prometheus gauge for the number of guilds the bot
// belongs to.
func GuildsGauge(config *Config) prometheus.Gauge {
	return newGauge(config.Log, "guilds", "Total Guilds count")
}

// MembersGauge returns a Prometheus gauge for the number of members of the
// guilds the bot belongs to.
func MembersGauge(config *Config) prometheus.Gauge {
	return newGauge(config.Log, "members", "Total Members count")
}

func newCounter(log *slog.Logger, name, help string) prometheus.Counter {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Name:      name,
		Help:      help,
	})

	if !register(log, counter, name) {
		return nil
	}

	return counter
}

func newGauge(log *slog.Logger, name, help string) prometheus.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Name:      name,
		Help:      help,
	})

	if !register(log, gauge, name) {
		return nil
	}

	return gauge
}

func register(log *slog.Logger, collector prometheus.Collector, name string) bool {
	if err := prometheus.Register(collector); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		log.Error("Unable to register metric with Prometheus", "metric", name, "error", err)
		return false
	}

	return true
}
