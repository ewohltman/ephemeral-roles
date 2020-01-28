package server

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/pprof"

	stdLog "log"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

// New returns a new pre-configured server instance.
func New(log logging.Interface, port string) *http.Server {
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
			drainCloseRequest(log, r)
		},
	)

	return &http.Server{
		Addr:     ":" + port,
		Handler:  mux,
		ErrorLog: stdLog.New(log.WrappedLogger().WriterLevel(logrus.ErrorLevel), "", 0),
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
