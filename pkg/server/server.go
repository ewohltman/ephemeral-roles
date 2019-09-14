package server

import (
	"net/http"
	"net/http/pprof"

	stdLog "log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// New returns a new pre-configured server instance
func New(log *logrus.Logger, port string) *http.Server {
	mux := http.NewServeMux()

	// Expose Prometheus metrics
	mux.Handle("/metrics", promhttp.Handler())

	// Expose pprof metrics
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Default handler
	mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	return &http.Server{
		Addr:     ":" + port,
		Handler:  mux,
		ErrorLog: stdLog.New(log.WriterLevel(logrus.ErrorLevel), "", 0),
	}
}
