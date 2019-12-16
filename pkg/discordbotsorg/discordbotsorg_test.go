package discordbotsorg

import (
	"net/http"
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	client := &http.Client{}

	err := Update(client, "", "", -1)
	if err != nil {
		if !strings.HasSuffix(err.Error(), `{"error":"Unauthorized"}`) {
			t.Error(err)
		}
	}
}
