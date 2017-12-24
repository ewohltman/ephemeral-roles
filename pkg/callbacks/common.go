// Package callbacks is a collection of the callback functions used in response
// to events from Discord's Websocket (WS)API.  Common definitions across the
// package are contained within common.go
package callbacks

import (
	"os"

	"github.com/ewohltman/ephemeral-roles/pkg/logging"
)

var log = logging.Instance()

// BOT_NAME is the name of the bot
var BOT_NAME = os.Getenv("BOT_NAME")

// BOT_KEYWORD is the keyword message prefix the bot should watch for
var BOT_KEYWORD = os.Getenv("BOT_KEYWORD")

// ROLE_PREFIX is the prefix to add before ephemeral role names
var ROLE_PREFIX = os.Getenv("ROLE_PREFIX")
