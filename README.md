# ephemeral-roles

| <a href="https://discordapp.com/api/oauth2/authorize?client_id=392419127626694676&permissions=268435456&scope=bot"><img src="https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/logo_Testa_anatomica_(1854)_-_Filippo_Balbi.jpg" width="100"></a><br/> ![build](https://github.com/ewohltman/ephemeral-roles/workflows/build/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ewohltman/ephemeral-roles)](https://goreportcard.com/report/github.com/ewohltman/ephemeral-roles) [![Coverage Status](https://coveralls.io/repos/github/ewohltman/ephemeral-roles/badge.svg?branch=master)](https://coveralls.io/github/ewohltman/ephemeral-roles?branch=master) [![GoDoc](https://godoc.org/github.com/ewohltman/ephemeral-roles/pkg?status.svg)](https://godoc.org/github.com/ewohltman/ephemeral-roles/pkg) ![Discord](https://img.shields.io/discord/393496906992713746?label=dev%20chat&logo=discord) |
| :------: |

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
| ![Ephemeral Roles action example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/action.gif) |
| ![Ephemeral Roles static example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/static.png) |
| ![Ephemeral Roles example role list](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/roles.png) |

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

When a member changes or disconnects from voice channels, even across Discord
servers, `Ephemeral Roles` will account for the change and automatically
revoke/reissue *ephemeral roles* as appropriate.

----

## Monitoring

A **[Prometheus](https://prometheus.io/)** and **[Grafana](https://grafana.com/)** instance have been set up to monitor `Ephemeral Roles` metrics.

Prometheus metrics are exposed via a `/metrics` HTTP end-point.

| [ephemeral-roles.net](http://ephemeral-roles.net/d/OqANQqtiz/ephemeral-roles-metrics?orgId=1&refresh=5s) |
| :------: |
| <a href="http://botmon.ephemeral-roles.com"><img src="https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/bot-metrics.png"></a> |

----

## Contributing to the project

Contributions are very welcome, however please follow the guidelines below:

* Open an issue describing the bug or enhancement
* Fork the `develop` branch and make your changes
  * Try to match current naming conventions as closely as possible
  * Try to keep changes small and incremental with appropriate new unit tests
* Create a Pull Request with your changes against the `develop` branch
