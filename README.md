# ephemeral-roles
Development branch: [![Travis CI](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)

`ephemeral-roles` is a Discord bot for managing ephemeral roles based upon user
voice channel presence.

## How does it work?

When the bot [Ephemeral Roles](https://discordapp.com/oauth2/authorize?&client_id=392419127626694676&scope=bot&permissions=0)
is invited to your Discord server, it will watch for changes to your voice
channels and when detecting a user joining a channel automatically assign them
a role associated with the channel.

In this way, it's easier to see at a glance who is available for chatting and
under which channel. 

When the user changes voice channels, even across servers, the bot will account
for it and will reissue/revoke roles as appropriate.  Upon disconnect from all
voice channels, all ephemeral roles are revoked.

## Rolling your own
 
In order to run this, you will need to define the following environment
variables:

```
LOG_LEVEL=info # Supported: debug, info, warn, error, fatal, panic
DISCORDRUS_WEBHOOK_URL= # Discord Webhook URL for bot logging
EPH_BOT_TOKEN= # Discord Bot Token
EPH_BOT_KEYWORD=!eph # Keyphrase to monitor incomming messages to begin with
EPH_CHANNEL_PREFIX=~ # Prefix to put before ephemeral channels to stand out 
EPH_CHANNEL_COLOR_HEX2DEC=16753920 # RGB color in hex to dec
PORT=8080 # port to bind for local HTTP server
```