package tracer_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const jaegerServiceName = "ephemeral-roles"

func TestRoundTripperFunc_RoundTrip(t *testing.T) {
	t.Parallel()

	reqBodyContent := []byte("Test message")
	reqBody := bytes.NewReader(reqBodyContent)

	mirrorRT := mock.NewMirrorRoundTripper()

	respBodyContent, err := doRoundTrip(mirrorRT, reqBody)
	require.NoError(t, err)

	assert.Equal(t, reqBodyContent, respBodyContent)
}

func TestNew(t *testing.T) {
	t.Parallel()

	testTracer, closer, err := newTestTracer()
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, closer.Close())
	}()

	require.NotNil(t, testTracer)
}

func TestRoundTripper(t *testing.T) {
	t.Parallel()

	jaegerTracer, closer, err := newTestTracer()
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, closer.Close())
	}()

	_, err = doRoundTrip(tracer.RoundTripper(jaegerTracer, "", mock.NewMirrorRoundTripper()), nil)
	require.NoError(t, err)
}

func doRoundTrip(roundTripper http.RoundTripper, reqBody io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "", reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating test request: %w", err)
	}

	resp, err := roundTripper.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("error performing round trip: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading test response body: %w", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing test response body: %w", err)
	}

	return respBody, nil
}

func newTestTracer() (opentracing.Tracer, io.Closer, error) {
	return tracer.New(jaegerServiceName)
}
