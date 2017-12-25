package discordBotsOrg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

const discordBotsURL = "https://discordbots.org"

var log = logging.Instance()

type serverUpdate struct {
	ServerCount int `json:"server_count"`
}

// Update POSTs a server_count update to discordbots.org
func Update(token string, botID string, serverCount int) {
	botID, found := os.LookupEnv("BOT_ID")
	if !found || botID == "" {
		log.WithField("warn", "BOT_ID not defined in environment variables").
			Errorf("Cannot POST updates to " + discordBotsURL)

		return
	}

	updateURL, err := url.Parse(discordBotsURL + "/api/bots/" + botID + "/stats")
	if err != nil {
		log.WithError(err).Errorf("Cannot POST updates to " + discordBotsURL)

		return
	}

	update := &serverUpdate{
		ServerCount: serverCount,
	}

	updateJSON, err := json.Marshal(update)
	if err != nil {
		log.WithError(err).Errorf("Error marshaling JSON body")

		return
	}

	client := &http.Client{}

	headers := make(http.Header)
	headers.Add("Authorization", token)
	headers.Add("Content-Type", "application/json")
	headers.Add("Content-Length", strconv.Itoa(len(updateJSON)))

	req := &http.Request{
		Method:        http.MethodPost,
		URL:           updateURL,
		Header:        headers,
		Body:          ioutil.NopCloser(bytes.NewReader(updateJSON)),
		ContentLength: int64(len(updateJSON)),
	}

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Errorf("Error updating " + discordBotsURL)

		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Errorf("Error reading response from " + discordBotsURL)

		return
	}

	if string(bodyBytes) != "{}" {
		log.WithField("response", string(bodyBytes)).Warnf("Abnormal " + discordBotsURL + " response")
	}
}
