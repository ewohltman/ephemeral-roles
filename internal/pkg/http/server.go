package http

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"sort"

	stdLog "log"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

const (
	metricsEndpoint = "/metrics"

	pprofIndexEndpoint   = "/debug/pprof/"
	pprofCmdlineEndpoint = "/debug/pprof/cmdline"
	pprofProfileEndpoint = "/debug/pprof/profile"
	pprofSymbolEndpoint  = "/debug/pprof/symbol"
	pprofTraceEndpoint   = "/debug/pprof/trace"

	guildsEndpoint = "/guilds"
	rootEndpoint   = "/"
)

type sortableGuild struct {
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

type sortableGuilds []sortableGuild

// Len returns the length of guilds to satisfy the sort.Interface interface.
func (guilds sortableGuilds) Len() int {
	return len(guilds)
}

// Less returns whether the element i is less than element j to satisfy the
// sort.Interface interface.
func (guilds sortableGuilds) Less(i, j int) bool {
	return guilds[i].MemberCount < guilds[j].MemberCount
}

// Swap swaps the elements i and j in the slice to satisfy the sort.Interface
// interface.
func (guilds sortableGuilds) Swap(i, j int) {
	guilds[i], guilds[j] = guilds[j], guilds[i]
}

// NewServer returns a new pre-configured *http.Server..
func NewServer(log logging.Interface, session *discordgo.Session, port string) *http.Server {
	mux := http.NewServeMux()

	mux.Handle(metricsEndpoint, promhttp.Handler())

	mux.HandleFunc(pprofIndexEndpoint, pprof.Index)
	mux.HandleFunc(pprofCmdlineEndpoint, pprof.Cmdline)
	mux.HandleFunc(pprofProfileEndpoint, pprof.Profile)
	mux.HandleFunc(pprofSymbolEndpoint, pprof.Symbol)
	mux.HandleFunc(pprofTraceEndpoint, pprof.Trace)

	mux.HandleFunc(guildsEndpoint, guildsHandler(log, session))
	mux.HandleFunc(rootEndpoint, rootHandler(log))

	errorLog := stdLog.New(log.WrappedLogger().WriterLevel(logrus.ErrorLevel), "", 0)

	return &http.Server{
		Addr:     "0.0.0.0:" + port,
		Handler:  mux,
		ErrorLog: errorLog,
	}
}

func guildsHandler(log logging.Interface, session *discordgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer drainCloseRequest(log, r)

		sortedGuilds := make(sortableGuilds, len(session.State.Guilds))

		for i, guild := range session.State.Guilds {
			sortedGuilds[i] = sortableGuild{
				Name:        guild.Name,
				MemberCount: guild.MemberCount,
			}
		}

		sort.Sort(sort.Reverse(sortedGuilds))

		sortedGuildsJSON, err := json.MarshalIndent(sortedGuilds, "", "    ")
		if err != nil {
			log.WithError(err).Errorf("Error marshaling sorted guilds to JSON")
			return
		}

		_, err = w.Write(sortedGuildsJSON)
		if err != nil {
			log.WithError(err).Errorf("Error writing sorted guilds response")
			return
		}
	}
}

func rootHandler(log logging.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		drainCloseRequest(log, r)
	}
}

func drainCloseRequest(log logging.Interface, r *http.Request) {
	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.WithError(err).Warn("Internal HTTP server error draining request body")
	}

	err = r.Body.Close()
	if err != nil {
		log.WithError(err).Warn("Internal HTTP server error closing request body")
	}
}
