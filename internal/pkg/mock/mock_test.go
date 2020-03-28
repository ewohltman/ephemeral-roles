package mock

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger()

	if log == nil {
		t.Fatal("unexpected nil Logger")
	}
}

func TestLogger_WrappedLogger(t *testing.T) {
	log := NewLogger().WrappedLogger()

	if log == nil {
		t.Fatal("unexpected nil wrapped *logrus.Logger")
	}
}

func TestLogger_UpdateLevel(t *testing.T) {
	NewLogger().UpdateLevel("info")
}

func TestNewMirrorRoundTripper(t *testing.T) {
	mirror := NewMirrorRoundTripper()

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

func TestNewSession(t *testing.T) {
	session, err := NewSession()
	if err != nil {
		t.Fatal(err)
	}

	defer SessionClose(t, session)

	_, err = session.User(TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Guild(TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildMember(TestGuild, TestUser)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoles(TestGuild)
	if err != nil {
		t.Error(err)
	}

	_, err = session.Channel(TestChannel)
	if err != nil {
		t.Error(err)
	}

	_, err = session.GuildRoleCreate(TestGuild)
	if err != nil {
		t.Error(err)
	}
}
