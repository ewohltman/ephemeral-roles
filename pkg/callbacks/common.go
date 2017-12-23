// Package callbacks is a collection of the callback functions used in response
// to events from Discord's Websocket (WS)API.  Common definitions across the
// package are contained within common.go
package callbacks

import (
	"os"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

var log = logging.Instance()

var BOT_NAME = os.Getenv("EPH_BOT_NAME")
var BOT_KEYWORD = os.Getenv("EPH_BOT_KEYWORD")
var BOT_ROLE_PREFIX = os.Getenv("EPH_ROLE_PREFIX")
