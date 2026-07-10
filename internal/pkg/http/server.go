package http

import (
	"cmp"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"slices"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Supported endpoints.
const (
	RootEndpoint   = "/"
	GuildsEndpoint = "/guilds"
	ReadyzEndpoint = "/readyz"
)

const (
	readHeaderTimeout = 3 * time.Second

	metricsEndpoint      = "/metrics"
	pprofIndexEndpoint   = "/debug/pprof/"
	pprofCmdlineEndpoint = "/debug/pprof/cmdline"
	pprofProfileEndpoint = "/debug/pprof/profile"
	pprofSymbolEndpoint  = "/debug/pprof/symbol"
	pprofTraceEndpoint   = "/debug/pprof/trace"
)

// SortableGuild is a representation of a guild that can be sorted by member
// count.
type SortableGuild struct {
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

// SortableGuilds is a slice of SortableGuild structs.
type SortableGuilds []SortableGuild

// NewServer returns a new pre-configured *http.Server..
func NewServer(log *slog.Logger, client *bot.Client, port string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc(RootEndpoint, rootHandler())
	mux.HandleFunc(GuildsEndpoint, guildsHandler(log, client))
	mux.HandleFunc(ReadyzEndpoint, readyzHandler(client))
	mux.HandleFunc(pprofIndexEndpoint, pprof.Index)
	mux.HandleFunc(pprofCmdlineEndpoint, pprof.Cmdline)
	mux.HandleFunc(pprofProfileEndpoint, pprof.Profile)
	mux.HandleFunc(pprofSymbolEndpoint, pprof.Symbol)
	mux.HandleFunc(pprofTraceEndpoint, pprof.Trace)
	mux.Handle(metricsEndpoint, promhttp.Handler())

	return &http.Server{
		Addr:              "0.0.0.0:" + port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}
}

func rootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		_, _ = w.Write(nil)
	}
}

// readyzHandler reports 200 once every shard this process manages has
// completed its gateway handshake and reached gateway.StatusReady, and 503
// otherwise. Kubernetes uses this to gate the StatefulSet's OrderedReady
// rollout: each pod is one shard, and disgo's IdentifyRateLimiter only spaces
// IDENTIFY calls out within a single process, so without this gate multiple
// pods starting up in quick succession can IDENTIFY within the same window
// and have Discord invalidate the colliding sessions.
func readyzHandler(client *bot.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()

		if !client.HasShardManager() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		for shard := range client.ShardManager.Shards() {
			if shard.Status() != gateway.StatusReady {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		_, _ = w.Write(nil)
	}
}

func guildsHandler(log *slog.Logger, client *bot.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		}()

		sortedGuilds := make(SortableGuilds, 0, client.Caches.GuildsLen())

		for guild := range client.Caches.Guilds() {
			sortedGuilds = append(sortedGuilds, SortableGuild{
				Name:        guild.Name,
				MemberCount: guild.MemberCount,
			})
		}

		slices.SortFunc(sortedGuilds, func(a, b SortableGuild) int {
			return cmp.Compare(b.MemberCount, a.MemberCount)
		})

		sortedGuildsJSON, err := json.MarshalIndent(sortedGuilds, "", "    ")
		if err != nil {
			log.Error("Error marshaling sorted guilds to JSON", "error", err)
			return
		}

		_, err = w.Write(sortedGuildsJSON)
		if err != nil {
			log.Error("Error writing sorted guilds response", "error", err)
			return
		}
	}
}
