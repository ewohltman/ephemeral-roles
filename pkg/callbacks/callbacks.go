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

// **** BEGIN: Role organization **********************************************
/*
// orderedRoles is a custom type for role organization
type orderedRoles []*discordgo.Role

// (oR orderedRoles) String satisfies the fmt.Stringer interface
func (oR orderedRoles) String() string {
	bufStr := ""

	for i := 0; i < len(oR); i++ {
		bufStr = fmt.Sprintf(
			"%s\nindex: %d, position: %d, name: %s",
			bufStr,
			i,
			oR[i].Position,
			oR[i].Name,
		)
	}

	return bufStr
}

func (oR orderedRoles) swap(i, j int) {
	oR[i].Position, oR[j].Position = oR[j].Position, oR[i].Position

	oR[i], oR[j] = oR[j], oR[i]
}
*/
