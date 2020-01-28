package discordbotsorg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	discordBotsURL   = "https://discordbots.org"
	discordbotsError = "discordbots.org integration error"
)

type serverUpdate struct {
	ServerCount int `json:"server_count"`
}

// Update POSTs a server_count update to discordbots.org
func Update(client *http.Client, token, botID string, serverCount int) error {
	req, err := buildRequest(botID, token, serverCount)
	if err != nil {
		return fmt.Errorf(
			discordbotsError+": error building request for ("+discordBotsURL+"): %w",
			err,
		)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf(
			discordbotsError+": error updating ("+discordBotsURL+"): %w",
			err,
		)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(
			discordbotsError+": error reading response from ("+discordBotsURL+"): %w",
			err,
		)
	}

	body := string(bodyBytes)
	if body != "{}" {
		return fmt.Errorf(
			discordbotsError+": abnormal response from ("+discordBotsURL+"): %s",
			body,
		)
	}

	return nil
}

func buildRequest(botID, token string, serverCount int) (*http.Request, error) {
	rawString := discordBotsURL + "/api/bots/" + botID + "/stats"

	updateURL, err := url.Parse(rawString)
	if err != nil {
		return nil, err
	}

	update := &serverUpdate{
		ServerCount: serverCount,
	}

	updateJSON, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

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

	return req, nil
}
