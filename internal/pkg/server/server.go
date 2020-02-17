// Package server provides an HTTP server implementation with handlers to
// expose Prometheus metrics.
package server

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

func (guilds sortableGuilds) Less(i, j int) bool {
	return guilds[i].MemberCount < guilds[j].MemberCount
}

// New returns a new pre-configured server instance.
func New(log logging.Interface, session *discordgo.Session, port string) *http.Server {
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

func guildsHandler(log logging.Interface, session *discordgo.Session) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer drainCloseRequest(log, r)

		sortedGuilds := make(sortableGuilds, len(session.State.Guilds))

		for i, guild := range session.State.Guilds {
			sortedGuilds[i] = sortableGuild{
				Name:        guild.Name,
				MemberCount: guild.MemberCount,
			}
		}

		sort.SliceStable(sortedGuilds, sortedGuilds.Less)

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

func rootHandler(log logging.Interface) func(http.ResponseWriter, *http.Request) {
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
