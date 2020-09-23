package mock_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
)

func TestNewMirrorRoundTripper(t *testing.T) {
	mirror := mock.NewMirrorRoundTripper()

	reqBodyContent := []byte("Test message")
	reqBody := bytes.NewReader(reqBodyContent)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "", reqBody)
	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	resp, err := mirror.RoundTrip(req)
	if err != nil {
		t.Fatalf("Error performing round trip: %s", err)
	}

	respBodyContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading test response body: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		t.Fatalf("Error closing test response body: %s", err)
	}

	if !reflect.DeepEqual(respBodyContent, reqBodyContent) {
		t.Fatalf(
			"Unexpected response body content. Expected: %s, Got: %s",
			string(reqBodyContent),
			string(respBodyContent),
		)
	}
}
