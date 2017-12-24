# ephemeral-roles
### A Discord bot for managing ephemeral roles based upon member voice channel presence

Development branch: [![Travis CI](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)
[![Discord Bots](https://discordbots.org/api/widget/status/392419127626694676.svg)](https://discordbots.org/bot/392419127626694676)

----

## Quickstart

1. [Invite](https://discordapp.com/oauth2/authorize?client_id=392419127626694676&scope=bot&permissions=268435456) the `Ephemeral Roles` bot to your Discord server
2. Ensure the new `Ephemeral Roles` role is at the top or as near as possible to the server's list of roles
4. Enjoy!

----

| Examples |
| :------: |
| ![Ephemeral Roles in action](https://media.giphy.com/media/3o6nUQ3e70R3uo5uzS/giphy.gif) |
| ![Ephemeral Roles example](https://i.imgur.com/RSHOAoz.png) |

----

## How does it work?

After the `Ephemeral Roles` bot is invited to your Discord server, it will
immediately start to watch for changes to your voice channels.  When a member
joins a channel, `Ephemeral Roles` will automatically assign that member an
*ephemeral role* associated with the channel.  If the *ephemeral role* doesn't
exist yet, `Ephemeral Roles` will create it.

By having your members auto-sorted into *ephemeral roles* in your member list,
it's clear to see who are available for chatting and the channels they are in
because the member list in Discord groups members together by role.

When a member changes voice channels, even across Discord servers,
`Ephemeral Roles` will account for the change and automatically revoke/reissue
*ephemeral roles* as appropriate.  Upon a member disconnecting from all voice channels,
`Ephemeral Roles` will revoke all *ephemeral roles*.

----

## Rolling your own locally
 
In order to run this locally, you will need to define the following environment
variables.

Required:
```
EPH_BOT_TOKEN= # Discord Bot Token
EPH_BOT_NAME= # Discord Bot Name.  I use this to differentiate "dev" vs "prod" bots
EPH_BOT_KEYWORD=![keyword] # Keyphrase to monitor incomming messages to begin with
EPH_ROLE_PREFIX={[keyword]} # Prefix to put before ephemeral channels to stand out 
EPH_CHANNEL_COLOR_HEX2DEC=16753920 # RGB color in hex to dec for the ephemeral roles
```

Optional:
```
PORT=8080 [optional] # port to bind for local HTTP server
LOG_LEVEL=info [optional] # Supported: debug, info, warn, error, fatal, panic.  Default: info
LOG_TIMEZONE_LOCATION=UTC [optional] # time.Location strings, e.g. "America/New_York".  Default: runtime time.Local
```

Optional integration with [discordrus](https://github.com/kz/discordrus)
```
DISCORDRUS_WEBHOOK_URL=[optional] # Webhook URL for discordrus bot logging to Discord integration
```

## Dependency Graph

![Dependency graph](https://github.com/ewohltman/ephemeral-roles/blob/develop/dep_status_visual.png)
