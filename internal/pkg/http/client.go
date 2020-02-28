package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
)

const contextTimeout = 20 * time.Second

// NewClient returns a new preconfigured *http.Client.
func NewClient(transport http.RoundTripper, jaegerTracer opentracing.Tracer) *http.Client {
	client := &http.Client{
		Transport: transport,
	}

	SetTransport(client, jaegerTracer)

	return client
}

// SetTransport takes an *http.Client and sets its Transport to a custom
// RoundTripper, wrapping an existing *http.Transport if already set or
// allocating a new *http.Transport if not already set.
func SetTransport(client *http.Client, jaegerTracer opentracing.Tracer) {
	if client.Transport != nil {
		transport, ok := client.Transport.(*http.Transport)
		if ok {
			client.Transport = roundTripperWithContext(transport)
			return
		}
	}

	transport := &http.Transport{}

	client.Transport = roundTripperWithTracer(
		jaegerTracer, roundTripperWithContext(
			transport,
		),
	)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface.
func (rt roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

func roundTripperWithTracer(jaegerTracer opentracing.Tracer, next http.RoundTripper) roundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		if jaegerTracer == nil {
			return next.RoundTrip(r)
		}

		carrier := opentracing.HTTPHeadersCarrier(r.Header)

		spanContext, err := jaegerTracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			if !errors.Is(err, opentracing.ErrSpanContextNotFound) {
				return nil, err
			}

			span := jaegerTracer.StartSpan(
				r.URL.String(),
				opentracing.StartTime(time.Now()),
			)

			spanContext = span.Context()
		}

		err = jaegerTracer.Inject(spanContext, opentracing.HTTPHeaders, carrier)
		if err != nil {
			return nil, err
		}

		return next.RoundTrip(r)
	}
}

func roundTripperWithContext(next http.RoundTripper) roundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		if r.Context() == context.Background() {
			ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
			defer cancelCtx()

			r = r.Clone(ctx)
		}

		return next.RoundTrip(r)
	}
}
