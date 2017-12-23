# ephemeral-roles
Development branch: [![Travis CI](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)

`ephemeral-roles` is a Discord bot for managing ephemeral roles based upon user
voice channel presence.

![Ephemeral Roles in action](https://media.giphy.com/media/3o6nUQ3e70R3uo5uzS/giphy.gif)

----

## Quickstart

1. Invite the `Ephemeral Roles` bot to your Discord server
    1. [Click here to invite](https://discordapp.com/oauth2/authorize?client_id=392419127626694676&scope=bot&permissions=268435456)
2. Ensure the new `Ephemeral Roles` role is at or as near as possible to the top of the server's list of roles
4. Enjoy!

----

## How does it work?

When the `Ephemeral Roles` bot is invited to your Discord server, it will watch
for changes to your voice channels.  When detecting a user joining a channel,
`Ephemeral Roles` will automatically assign that user an ephemeral role
associated with the channel. 

In this way, it's easier to see at a glance who is available for chatting and
under which channel because the member list in Discord groups members together
by role.

When the user changes voice channels, even across Discord servers,
`Ephemeral Roles` will account for the change and automatically revoke/reissue
ephemeral roles as appropriate.  Upon a user disconnecting from all voice channels,
`Ephemeral Roles` will revoke all ephemeral roles.

----

## Rolling your own
 
In order to run this locally, you will need to define the following environment
variables:

```
PORT=8080 # port to bind for local HTTP server
LOG_LEVEL=info # Supported: debug, info, warn, error, fatal, panic
LOG_TIMEZONE_LOCATION=UTC [optional] # time.Location strings, e.g. "America/New_York".  Default: runtime time.Local
EPH_BOT_TOKEN= # Discord Bot Token
EPH_BOT_NAME= # Discord Bot Name.  I use this to differentiate "dev" vs "prod" bots
EPH_BOT_KEYWORD=![keyword] # Keyphrase to monitor incomming messages to begin with
EPH_ROLE_PREFIX={[keyword]} # Prefix to put before ephemeral channels to stand out 
EPH_CHANNEL_COLOR_HEX2DEC=16753920 # RGB color in hex to dec for the ephemeral roles
DISCORDRUS_WEBHOOK_URL=[optional] # Webhook URL for discordrus bot logging to Discord
```