// Package callbacks is a collection of the callback functions used in response
// to events from Discord's Websocket (WS)API.  Common definitions across the
// package are contained within common.go
package callbacks

import (
	"os"

	"github.com/ewohltman/discordEphemeralRolesProject/pkg/logging"
)

var log = logging.Instance()
var botKeyphrase = os.Getenv("DERP_BOT_KEYWORD")
var botChannelPrefix = os.Getenv("DERP_CHANNEL_PREFIX")
