package http

import (
	"context"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const contextTimeout = 20 * time.Second

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (rt roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// NewClient returns a new preconfigured *http.Client.
func NewClient(transport http.RoundTripper, jaegerTracer opentracing.Tracer, instanceName string) *http.Client {
	client := &http.Client{Transport: transport}

	setTransport(client, jaegerTracer, instanceName)

	return client
}

func setTransport(client *http.Client, jaegerTracer opentracing.Tracer, instanceName string) {
	if client.Transport == nil {
		client.Transport = &http.Transport{}
	}

	client.Transport = roundTripperWithTracer(
		jaegerTracer,
		instanceName,
		roundTripperWithContext(client.Transport),
	)
}

func roundTripperWithContext(next http.RoundTripper) http.RoundTripper {
	return roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			if req.Context() != context.Background() {
				return next.RoundTrip(req)
			}

			ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
			defer cancelCtx()

			return next.RoundTrip(req.Clone(ctx))
		},
	)
}

func roundTripperWithTracer(jaegerTracer opentracing.Tracer, instanceName string, next http.RoundTripper) http.RoundTripper {
	return tracer.RoundTripper(jaegerTracer, instanceName, next)
}
