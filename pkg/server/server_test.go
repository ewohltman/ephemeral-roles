package server

import "testing"

func TestNew(t *testing.T) {
	testServer := New("8080")
	if testServer == nil {
		t.Errorf("Failed creating new internal HTTP server")
	}
}

func TestMonitorGuildsUpdate(t *testing.T) {

}
