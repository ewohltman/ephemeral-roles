// Package tracer provides functionality for using Jaeger and OpenTracing for
// instrumenting HTTP requests to collect metrics.
package tracer

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
)

const (
	samplerProbability = 1
	samplerType        = jaeger.SamplerTypeConst
)

// RoundTripperFunc allows functions to satisfy the http.RoundTripper
// interface.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface.
func (rt RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

type jaegerLogger struct {
	log logging.Interface
}

// Infof satisfies the jaeger.Logger interface by delegating to the wrapped
// logging.Interface Error method.
func (jaegerLog *jaegerLogger) Infof(msg string, args ...interface{}) {
	jaegerLog.log.Debugf(msg, args...)
}

// Error satisfies the jaeger.Logger interface by delegating to the wrapped
// logging.Interface Error method.
func (jaegerLog *jaegerLogger) Error(msg string) {
	jaegerLog.log.Debug(msg)
}

// New returns a new opentracing.Tracer and io.Closer to be used for
// instrumenting HTTP requests to collect metrics.
func New(log logging.Interface, serviceName string) (opentracing.Tracer, io.Closer, error) {
	cfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  samplerType,
			Param: samplerProbability,
		},
		Reporter: &config.ReporterConfig{
			BufferFlushInterval: time.Second,
		},
	}

	tracer, closer, err := cfg.NewTracer(
		config.Logger(&jaegerLogger{log: log}),
		config.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("could not initialize Jaeger tracer: %w", err)
	}

	return tracer, closer, nil
}

// NewSpan returns a new opentracing.Span. If parentSpan is not nil, it will be
// a child of parentSpan.
func NewSpan(jaegerTracer opentracing.Tracer, parentSpanContext opentracing.SpanContext, operationName string) opentracing.Span {
	if parentSpanContext != nil {
		return jaegerTracer.StartSpan(
			operationName,
			opentracing.ChildOf(parentSpanContext),
		)
	}

	return jaegerTracer.StartSpan(operationName)
}

// RoundTripper is http.RoundTripper middleware to add Jaeger tracing to all
// HTTP requests.
func RoundTripper(jaegerTracer opentracing.Tracer, parentSpanContext opentracing.SpanContext, next http.RoundTripper) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		if jaegerTracer == nil {
			return next.RoundTrip(req)
		}

		spanFromParent := NewSpan(jaegerTracer, parentSpanContext, req.URL.String())
		defer spanFromParent.Finish()

		carrier := opentracing.HTTPHeadersCarrier(req.Header)

		err := jaegerTracer.Inject(spanFromParent.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			return nil, err
		}

		return next.RoundTrip(req)
	}
}
