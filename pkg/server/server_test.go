package server

import (
	"testing"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	log := logging.New()
	log.SetLevel(logrus.FatalLevel)

	testServer := New(log, "8080")
	if testServer == nil {
		t.Errorf("Failed creating new internal HTTP server")
	}
}
