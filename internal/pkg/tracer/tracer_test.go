package tracer_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/opentracing/opentracing-go"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const jaegerServiceName = "ephemeral-roles"

func TestRoundTripperFunc_RoundTrip(t *testing.T) {
	reqBodyContent := []byte("Test message")
	reqBody := bytes.NewReader(reqBodyContent)

	mirrorRT := mock.NewMirrorRoundTripper()

	respBodyContent, err := doRoundTrip(mirrorRT, reqBody)
	if err != nil {
		t.Fatalf("Error performing round trip: %s", err)
	}

	if !reflect.DeepEqual(respBodyContent, reqBodyContent) {
		t.Fatalf(
			"Unexpected response body content. Expected: %s, Got: %s",
			string(reqBodyContent),
			string(respBodyContent),
		)
	}
}

func TestNew(t *testing.T) {
	testTracer, closer, err := newTestTracer()
	if err != nil {
		t.Fatalf("Error creating test tracer: %s", err)
	}

	defer func() {
		closeErr := closer.Close()
		if closeErr != nil {
			t.Errorf("Error closing test tracer: %s", err)
		}
	}()

	if testTracer == nil {
		t.Fatal("Unexpected nil tracer")
	}
}

func TestRoundTripper(t *testing.T) {
	jaegerTracer, closer, err := newTestTracer()
	if err != nil {
		t.Fatalf("Error creating test tracer: %s", err)
	}

	defer func() {
		closeErr := closer.Close()
		if closeErr != nil {
			t.Errorf("Error closing test tracer: %s", err)
		}
	}()

	_, err = doRoundTrip(tracer.RoundTripper(jaegerTracer, "", mock.NewMirrorRoundTripper()), nil)
	if err != nil {
		t.Fatalf("Error performing round trip: %s", err)
	}
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
