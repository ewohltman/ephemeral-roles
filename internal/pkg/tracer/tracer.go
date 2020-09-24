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
)

const (
	samplerProbability = 1
	samplerType        = jaeger.SamplerTypeConst

	operationName = "HTTP request"
)

// RoundTripperFunc allows functions to satisfy the http.RoundTripper
// interface.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface.
func (rt RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// New returns a new opentracing.Tracer and io.Closer to be used for
// instrumenting HTTP requests to collect metrics.
func New(serviceName string) (opentracing.Tracer, io.Closer, error) {
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
		config.Logger(jaeger.NullLogger),
		config.Metrics(metrics.NullFactory),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("could not initialize Jaeger tracer: %w", err)
	}

	return tracer, closer, nil
}

// RoundTripper is http.RoundTripper middleware to add Jaeger tracing to all
// HTTP requests.
func RoundTripper(jaegerTracer opentracing.Tracer, instanceName string, next http.RoundTripper) RoundTripperFunc {
	return func(req *http.Request) (*http.Response, error) {
		if jaegerTracer == nil {
			return next.RoundTrip(req)
		}

		span, traceCtx := opentracing.StartSpanFromContextWithTracer(req.Context(), jaegerTracer, operationName)
		defer span.Finish()

		span.SetTag("url", req.URL.Path)
		span.SetTag("instance", instanceName)
		span.SetTag("method", req.Method)

		resp, err := next.RoundTrip(req.Clone(traceCtx))
		if err != nil {
			span.SetTag("error", err.Error())
			return resp, err
		}

		span.SetTag("response", resp.StatusCode)

		return resp, err
	}
}
