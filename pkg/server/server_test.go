package server

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

const testPort = "8080"

func TestNew(t *testing.T) {
	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

	testServer := New(log, testPort)
	if testServer == nil {
		t.Fatal("Unexpected nil *http.Server")
	}

	stopChan := make(chan struct{})

	go func() {
		err := testServer.ListenAndServe()
		if err != nil {
			t.Fatalf("Error starting test server: %s", err)
		}

		<-stopChan

		err = testServer.Close()
		if err != nil {
			t.Errorf("Error closing test server: %s", err)
		}
	}()

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, "http://localhost:"+testPort, bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("Error creating test request: %s", err)
	}

	_, err = client.Do(req)
	if err != nil {
		t.Fatalf("Error performing test request: %s", err)
	}

	close(stopChan)
}
