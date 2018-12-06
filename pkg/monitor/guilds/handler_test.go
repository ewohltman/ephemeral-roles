package guilds

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	HTTPHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	_, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected response code from /guilds")
	}
}
