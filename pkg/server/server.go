package server

import (
	"net/http"
	"net/http/pprof"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor/guilds"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var log = logging.Instance()

// S is the struct for the internal HTTP server
type S struct {
	Logger *logrus.Logger
	mux    *http.ServeMux
}

// (s *S) ServeHTTP satisfies the http.Handler interface
func (s *S) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// New returns a new pre-configured server instance
func New(port string, dgBotSession *discordgo.Session, token string, botID string) *http.Server {
	monitor.Start(dgBotSession, token, botID)

	return &http.Server{
		Addr: ":" + port,
		Handler: server(
			func(s *S) {
				s.Logger = log
			},
		),
	}
}

func server(options ...func(*S)) *S {
	s := &S{
		Logger: log,
		mux:    http.NewServeMux(),
	}

	for _, f := range options {
		f(s)
	}

	// List the guilds our bot is a member of
	s.mux.HandleFunc("/guilds", guilds.HTTPHandler)

	// Expose Prometheus metrics
	s.mux.Handle("/metrics", promhttp.Handler())

	// Expose pprof metrics
	s.mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	s.mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	s.mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	s.mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	s.mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// Default handler
	s.mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	return s
}
