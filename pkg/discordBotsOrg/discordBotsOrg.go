package discordBotsOrg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

const discordBotsURL = "https://discordbots.org"

type serverUpdate struct {
	ServerCount int `json:"server_count"`
}

// Update POSTs a server_count update to discordbots.org
func Update(token string, botID string, serverCount int) (string, error) {
	rawString := discordBotsURL + "/api/bots/" + botID + "/stats"

	updateURL, err := url.Parse(rawString)
	if err != nil {
		return "", errors.Wrap(
			err,
			"discordbots.org integration disabled: unable to parse ("+
				rawString+
				"): "+
				err.Error(),
		)
	}

	update := &serverUpdate{
		ServerCount: serverCount,
	}

	updateJSON, err := json.Marshal(update)
	if err != nil {
		return "", errors.Wrap(
			err,
			"discordbots.org integration: error marshaling JSON body: "+
				err.Error(),
		)
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
		return "", errors.Wrap(
			err,
			"discordbots.org integration: error updating ("+
				discordBotsURL+
				"): "+
				err.Error(),
		)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(
			err,
			"discordbots.org integration: error reading response from ("+
				discordBotsURL+
				"): "+
				err.Error(),
		)
	}

	return string(bodyBytes), nil
}
