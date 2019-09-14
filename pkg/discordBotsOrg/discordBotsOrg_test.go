package discordBotsOrg

import (
	"strings"
	"testing"
)

func TestUpdate(t *testing.T) {
	err := Update("", "", -1)
	if err != nil {
		if !strings.HasSuffix(err.Error(), `{"error":"Unauthorized"}`) {
			t.Error(err)
		}
	}
}
