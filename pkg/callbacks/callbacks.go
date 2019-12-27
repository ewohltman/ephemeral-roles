package callbacks

import (
	"bytes"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Log                     *logrus.Logger
	BotName                 string
	BotKeyword              string
	RolePrefix              string
	ReadyCounter            prometheus.Counter
	MessageCreateCounter    prometheus.Counter
	VoiceStateUpdateCounter prometheus.Counter
}

// DiscordAPIResponse is a receiving struct for API JSON responses
type DiscordAPIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type discordError struct {
	HTTPResponseMessage string
	APIResponse         *DiscordAPIResponse
	CustomMessage       string
}

// Error satisfies the error interface for discordError
func (dErr *discordError) Error() string {
	return "error from Discord API: " + dErr.String()
}

// String satisfies the fmt.Stringer interface for discordError
func (dErr *discordError) String() string {
	buf := bytes.NewBuffer([]byte{})

	if dErr.CustomMessage != "" {
		buf.Write([]byte("CustomMessage: " + dErr.CustomMessage + ", "))
	}

	buf.Write([]byte("HTTPResponseMessage: " + dErr.HTTPResponseMessage + ", "))
	buf.Write([]byte("APIResponse.Code: " + strconv.Itoa(dErr.APIResponse.Code) + ", "))
	buf.Write([]byte("APIResponse.Message: " + dErr.APIResponse.Message))

	return buf.String()
}
