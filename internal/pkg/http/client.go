package http

import (
	"context"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const contextTimeout = 30 * time.Second

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// NewClient returns a new preconfigured *http.Client.
func NewClient(transport http.RoundTripper, jaegerTracer opentracing.Tracer, parentCtx opentracing.SpanContext) *http.Client {
	client := &http.Client{
		Transport: transport,
	}

	SetTransport(client, jaegerTracer, parentCtx)

	return client
}

// SetTransport takes an *http.Client and wraps its Transport with RoundTripper
// middleware. If the *http.Client does not have an initial Transport, a new
// *http.Transport will be allocated for it.
func SetTransport(client *http.Client, jaegerTracer opentracing.Tracer, parentCtx opentracing.SpanContext) {
	transport := client.Transport

	if transport == nil {
		transport = &http.Transport{}
	}

	client.Transport = roundTripperWithTracer(jaegerTracer, parentCtx,
		roundTripperWithContext(
			transport,
		),
	)
}

func roundTripperWithTracer(jaegerTracer opentracing.Tracer, parentCtx opentracing.SpanContext, next http.RoundTripper) http.RoundTripper {
	return tracer.RoundTripper(jaegerTracer, parentCtx, next)
}

func roundTripperWithContext(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			if req.Context() == context.Background() {
				ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
				defer cancelCtx()

				req = req.Clone(ctx)
			}

			return next.RoundTrip(req)
		},
	)
}
