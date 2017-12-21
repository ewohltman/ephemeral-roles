# ephemeral-roles
Development branch: [![Travis CI](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)](https://travis-ci.org/ewohltman/ephemeral-roles.svg?branch=develop)

`ephemeral-roles` is a Discord bot for managing ephemeral roles based upon user
voice channel presence.

----

## Quickstart

1. Invite the `Ephemeral Roles` bot to your Discord server
    1. [Click here to invite](https://discordapp.com/oauth2/authorize?&client_id=392419127626694676&scope=bot&permissions=0)
2. Create a new role in your server specifically for `Ephemeral Roles`
    1. Ensure the new role is at or as near to the top of the server's list of roles as possible
    2. Enable the 'Manage Roles' permission for the new `Ephemeral Roles` role
3. Assign the new role to `Ephemeral Roles`
4. Enjoy

----

## How does it work?

When the `Ephemeral Roles` bot is invited to your Discord server, it will watch
for changes to your voice channels.  When detecting a user joining a channel,
`Ephemeral Roles` will automatically assign that user an ephemeral role
associated with the channel. 

In this way, it's easier to see at a glance who is available for chatting and
under which channel because the member list in Discord servers group members
together by role.

When the user changes voice channels, even across Discord servers,
`Ephemeral Roles` will account for the change and automatically revoke/reissue
ephemeral roles as appropriate.  Upon a user disconnecting from all voice channels,
`Ephemeral Roles` all revoke all ephemeral roles.

----

## Rolling your own
 
In order to run this, you will need to define the following environment
variables:

```
LOG_LEVEL=info # Supported: debug, info, warn, error, fatal, panic
PORT=8080 # port to bind for local HTTP server
DISCORDRUS_WEBHOOK_URL= # Discord Webhook URL for bot logging
EPH_BOT_TOKEN= # Discord Bot Token
EPH_BOT_KEYWORD=!eph # Keyphrase to monitor incomming messages to begin with
EPH_CHANNEL_PREFIX=~ # Prefix to put before ephemeral channels to stand out 
EPH_CHANNEL_COLOR_HEX2DEC=16753920 # RGB color in hex to dec for the ephemeral roles
```