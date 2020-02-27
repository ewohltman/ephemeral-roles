// Package tracer provides functionality for using Jaeger and OpenTracing for
// instrumenting HTTP requests to collect metrics.
package tracer

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

type jaegerLogger struct {
	log logging.Interface
}

// Infof satisfies the jaeger.Logger interface by delegating to the wrapped
// logging.Interface Error method.
func (jaegerLog *jaegerLogger) Infof(msg string, args ...interface{}) {
	jaegerLog.log.Infof(msg, args...)
}

// Error satisfies the jaeger.Logger interface by delegating to the wrapped
// logging.Interface Error method.
func (jaegerLog *jaegerLogger) Error(msg string) {
	jaegerLog.log.Error(msg)
}

// New returns a new opentracing.Tracer and io.Closer to be used for
// instrumenting HTTP requests to collect metrics.
func New(log logging.Interface, serviceName string, sampleProbability float64) (opentracing.Tracer, io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: serviceName,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: sampleProbability,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(&jaegerLogger{log: log}),
		jaegercfg.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("could not initialize Jaeger tracer: %w", err)
	}

	return tracer, closer, nil
}
