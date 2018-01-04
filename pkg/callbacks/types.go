package callbacks

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/ewohltman/discordgo"
)

// **** BEGIN: API response structs and associated methods ********************

// DiscordAPIResponse is a receiving struct for API JSON responses
type DiscordAPIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// (dAR *DiscordAPIResponse) String satisfies the fmt.Stringer interface for field names in logs
func (dAR *DiscordAPIResponse) String() string {
	return fmt.Sprintf("Code: %d, Message: %s", dAR.Code, dAR.Message)
}

// discordError is a convenience struct for encapsulating API error responses
// for logging
type discordError struct {
	HTTPResponseMessage string
	APIResponse         *DiscordAPIResponse
	CustomMessage       string
}

// (dErr *discordError) Error satisfies the error interface
func (dErr *discordError) Error() string {
	return "error from Discord API: " + dErr.String()
}

// (dErr *discordError) String satisfies the fmt.Stringer interface
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
