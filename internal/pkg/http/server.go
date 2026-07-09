package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"sort"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Supported endpoints.
const (
	RootEndpoint   = "/"
	GuildsEndpoint = "/guilds"
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

// Len returns the length of guilds to satisfy the sort.Interface interface.
func (guilds SortableGuilds) Len() int {
	return len(guilds)
}

// Less returns whether the element i is less than element j to satisfy the
// sort.Interface interface.
func (guilds SortableGuilds) Less(i, j int) bool {
	return guilds[i].MemberCount < guilds[j].MemberCount
}

// Swap swaps the elements i and j in the slice to satisfy the sort.Interface
// interface.
func (guilds SortableGuilds) Swap(i, j int) {
	guilds[i], guilds[j] = guilds[j], guilds[i]
}

// NewServer returns a new pre-configured *http.Server..
func NewServer(log *slog.Logger, client *bot.Client, port string) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc(RootEndpoint, rootHandler())
	mux.HandleFunc(GuildsEndpoint, guildsHandler(log, client))
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

		sort.Sort(sort.Reverse(sortedGuilds))

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
