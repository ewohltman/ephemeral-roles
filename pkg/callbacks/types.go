package callbacks

import (
	"bytes"
	"fmt"
	"sort"
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

// **** BEGIN: Channel organization *******************************************

// orderedChannels is a custom type for channel organization
type orderedChannels []*discordgo.Channel

// (oC orderedChannels) Len is to satisfy the sort.Interface interface
func (oC orderedChannels) Len() int {
	return len(oC)
}

// (oC orderedChannels) Less is to satisfy the sort.Interface interface
func (oC orderedChannels) Less(i, j int) bool {
	return oC[i].Position < oC[j].Position
}

// (oC orderedChannels) Swap is to satisfy the sort.Interface interface
func (oC orderedChannels) Swap(i, j int) {
	oC[i].Position, oC[j].Position = oC[j].Position, oC[i].Position
	oC[i], oC[j] = oC[j], oC[i]
}

// (oC orderedChannels) String satisfies the fmt.Stringer interface
func (oC orderedChannels) String() string {
	if !sort.IsSorted(oC) {
		sort.Stable(oC)
	}

	bufStr := ""

	for index, channel := range oC {
		bufStr = fmt.Sprintf(
			"%s\nindex: %d, position: %d, name: %s",
			bufStr,
			index,
			channel.Position,
			channel.Name,
		)
	}

	return bufStr
}

// (oC orderedChannels) voiceChannelsSort is a convenience method for returning
// the ordered voice channels in a generic orderedChannels
func (oC orderedChannels) voiceChannelsSort() (oVC orderedChannels) {
	oVC = make([]*discordgo.Channel, 0, len(oC))

	for _, channel := range oC {
		if channel.Type == discordgo.ChannelTypeGuildVoice {
			oVC = append(oVC, channel)
		}
	}

	sort.Stable(oVC)

	return
}

// **** BEGIN: Role organization **********************************************

// orderedRoles is a custom type for role organization
type orderedRoles []*discordgo.Role

// (oR orderedRoles) Len is to satisfy the sort.Interface interface
func (oR orderedRoles) Len() int {
	return len(oR)
}

// (oR orderedRoles) Less is to satisfy the sort.Interface interface
func (oR orderedRoles) Less(i, j int) bool {
	return oR[i].Position < oR[j].Position
}

// (oR orderedRoles) Swap is to satisfy the sort.Interface interface
func (oR orderedRoles) Swap(i, j int) {
	oR[i], oR[j] = oR[j], oR[i]
	oR[i].Position, oR[j].Position = oR[j].Position, oR[i].Position
}

// (oR orderedRoles) sort is a convenience method for sorting roles
func (oR orderedRoles) sort() orderedRoles {
	for index, role := range oR {
		if role.Name == "@everyone" { // @everyone should be the lowest
			if role.Position != 0 { // ...and it's not
				tmpPos := oR[0].Position
				oR[0].Position = oR[index].Position
				oR[index].Position = tmpPos
			}
		}

		if role.Name == BOTNAME { // Our bot role should be the highest
			if role.Position != len(oR)-1 { // ...and it's not
				tmpPos := oR[len(oR)-1].Position
				oR[len(oR)-1].Position = oR[index].Position
				oR[index].Position = tmpPos
			}
		}
	}

	sort.Stable(oR)

	return oR
}

// (oR orderedRoles) String satisfies the fmt.Stringer interface
func (oR orderedRoles) String() string {
	bufStr := ""

	for i := 0; i < len(oR); i++ {
		bufStr = fmt.Sprintf(
			"%s\nposition: %d, name: %s",
			bufStr,
			oR[i].Position,
			oR[i].Name,
		)
	}

	return bufStr
}
