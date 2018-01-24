# ephemeral-roles

| <a href="https://discordapp.com/api/oauth2/authorize?client_id=392419127626694676&permissions=268435456&scope=bot"><img src="https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/logo_Testa_anatomica_(1854)_-_Filippo_Balbi.jpg" width="100"></a> | [![Discord Bots Status](https://discordbots.org/api/widget/status/392419127626694676.svg)](https://discordbots.org/bot/ephemeral-roles) [![Discord Bots Servers](https://discordbots.org/api/widget/servers/392419127626694676.svg)](https://discordbots.org/bot/ephemeral-roles)<br/>[![Travis CI](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=master)](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ewohltman/ephemeral-roles)](https://goreportcard.com/report/github.com/ewohltman/ephemeral-roles)<br/>[![GoDoc](https://godoc.org/github.com/ewohltman/ephemeral-roles/pkg?status.svg)](https://godoc.org/github.com/ewohltman/ephemeral-roles/pkg) |
| :------: | :------: |

### A Discord bot for managing ephemeral roles based upon voice channel member presence

----

## Quickstart

1. Click on the `Ephemeral Roles` logo head above or use [this link](https://discordapp.com/api/oauth2/authorize?client_id=392419127626694676&permissions=268435456&scope=bot)
to invite `Ephemeral Roles` into your Discord server
    1. The 'Manage Roles' permission is required.  The invite link above
    provides that by automatically creating an appropriate role in your server
    for `Ephemeral Roles` 
2. Ensure the new role for `Ephemeral Roles` is at the top (or as near as
possible) to the server's list of roles
    1. If you're not sure how or why to do that, take a quick read over
    Discord's excellent [Role Management 101](https://support.discordapp.com/hc/en-us/articles/214836687-Role-Management-101) guide
3. Enjoy!

----

| Usage/Examples \(Orange roles below are automatically managed by `Ephemeral Roles`\) |
| :------: |
| ![Ephemeral Roles action example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/action.gif) |
| ![Ephemeral Roles static example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/static.png) |
| ![Ephemeral Roles example role list](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/res/roles.png) |

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
all voice channels `Ephemeral Roles` will revoke all *ephemeral roles*.

----

## Rolling your own locally
 
In order to run this locally, you will need to define the following environment
variables.

**Required:**
```
BOT_TOKEN= # Discord Bot Token
BOT_NAME= # Discord Bot Name
BOT_KEYWORD=![keyword] # Keyphrase to monitor incomming messages to begin with
ROLE_PREFIX={[keyword]} # Prefix to put before ephemeral channels to stand out
```

**Optional:**
```
ROLE_COLOR_HEX2DEC=16753920 # RGB color in hex to dec. Default: orange
PORT=8080 # Port to bind for local HTTP server. Default: 8080
LOG_LEVEL=info # Levels: debug, info, warn, error, fatal, panic. Default: info
LOG_TIMEZONE_LOCATION=UTC # "America/New_York". Default: runtime time.Local
```

**Optional integration with [discordrus](https://github.com/kz/discordrus):**
```
DISCORDRUS_WEBHOOK_URL= # Webhook URL for discordrus bot logging to Discord integration
```

**Optional integration with [discordbots.org](https://discordbots.org/):**
```
DISCORDBOTS_ORG_TOKEN= # Token from discordbots.org for POSTing updates
BOT_ID= # Discord Bot Client ID
```

----

## Dependencies

| [dep](https://github.com/golang/dep) Graph |
| :------: |
| ![Dependency graph](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/dep_status_visual.png) |
