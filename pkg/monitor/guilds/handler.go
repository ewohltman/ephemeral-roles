package guilds

import (
	"bytes"
	"net/http"
	"strconv"
)

// HTTPHandler is the function used to handle /guilds HTTP requests
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	buf := bytes.NewBuffer([]byte{})
	for _, guild := range cache.guildList {
		buf.Write([]byte(guild.Name + "\n"))
	}

	response := buf.Bytes()

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))

	_, err := w.Write(response)
	if err != nil {
		log.WithError(err).Errorf("Error writing /check HTTP response")
		return
	}
}
