package server

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

const testPort = "8080"

func TestNew(t *testing.T) {
	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

	testServer := New(log, testPort)
	if testServer == nil {
		t.Fatal("Unexpected nil *http.Server")
	}

	var serverErr error

	serverClosed := make(chan struct{})

	go func() {
		serverErr = testServer.ListenAndServe()

		close(serverClosed)
	}()

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, "http://localhost:"+testPort, bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error performing test request: %s", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			t.Errorf("Error closing test response body: %s", err)
		}
	}()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		t.Errorf("Error draining test response body: %s", err)
	}

	err = testServer.Close()
	if err != nil {
		t.Errorf("Error closing test server: %s", err)
	}

	<-serverClosed

	if !errors.Is(serverErr, http.ErrServerClosed) {
		t.Errorf("Test server error: %s", serverErr)
	}
}
