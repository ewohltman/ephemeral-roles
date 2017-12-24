# ephemeral-roles

<img src="https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/logo_Testa_anatomica_(1854)_-_Filippo_Balbi.jpg" width="100">

[![Discord Bots](https://discordbots.org/api/widget/status/392419127626694676.svg)](https://discordbots.org/bot/392419127626694676)
[![Travis CI](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=master)](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ewohltman/ephemeral-roles)](https://goreportcard.com/report/github.com/ewohltman/ephemeral-roles)
[![GoDoc](https://godoc.org/github.com/ewohltman/ephemeral-roles/pkg?status.svg)](https://godoc.org/github.com/ewohltman/ephemeral-roles/pkg)

### A Discord bot for managing ephemeral roles based upon voice channel member presence

----

## Quickstart

1. [Invite](https://discordapp.com/oauth2/authorize?client_id=392419127626694676&scope=bot&permissions=268435456) the `Ephemeral Roles` bot to your Discord server
2. Ensure the new `Ephemeral Roles` role is at the top (or as near as possible) to the server's list of roles
4. Enjoy!

----

| Usage |
| :------: |
| ![Ephemeral Roles action example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/action.gif) |
| ![Ephemeral Roles static example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/static.png) |

----

## How does it work?

After the `Ephemeral Roles` bot is invited to your Discord server, it
immediately starts to watch for changes to your voice channels.  When a member
joins a channel, `Ephemeral Roles` automatically assigns that member an
*ephemeral role* associated with the channel.  If the *ephemeral role* doesn't
exist yet, `Ephemeral Roles` will create it.

By having your members auto-sorted into *ephemeral roles* in your member list,
it's clear to see who are available for chatting and the channels they are in.
This is because `ephemeral-roles` leverages the Discord feature that the member
list in servers will group together members by role right out of the box.

When a member changes voice channels, even across Discord servers,
`Ephemeral Roles` will account for the change and automatically revoke/reissue
*ephemeral roles* as appropriate.  Finally, upon a member disconnecting from
all voice channels, `Ephemeral Roles` will revoke all *ephemeral roles*.

----

## Rolling your own locally
 
In order to run this locally, you will need to define the following environment
variables.

**Required:**
```
BOT_TOKEN= # Discord Bot Token
BOT_NAME= # Discord Bot Name.  I use this to differentiate "dev" vs "prod" bots
BOT_KEYWORD=![keyword] # Keyphrase to monitor incomming messages to begin with
ROLE_PREFIX={[keyword]} # Prefix to put before ephemeral channels to stand out
```

**Optional:**
```
ROLE_COLOR_HEX2DEC=16753920 # RGB color in hex to dec for the ephemeral roles.  Default: orange
PORT=8080 # Port to bind for local HTTP server.  Default: 8080
LOG_LEVEL=info # Supported: debug, info, warn, error, fatal, panic.  Default: info
LOG_TIMEZONE_LOCATION=UTC # time.Location strings, e.g. "America/New_York".  Default: runtime time.Local
```

**Optional integration with [discordrus](https://github.com/kz/discordrus):**
```
DISCORDRUS_WEBHOOK_URL= # Webhook URL for`discordrus bot logging to Discord integration
```

## Dependencies

| [dep](https://github.com/golang/dep) Graph |
| :------: |
| ![Dependency graph](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/dep_status_visual.png) |
