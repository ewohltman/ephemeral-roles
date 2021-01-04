# ephemeral-roles

| <a href="https://discordapp.com/api/oauth2/authorize?client_id=392419127626694676&permissions=268435456&scope=bot"><img src="https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/Testa_Anatomica-Filippo_Balbi.jpg" width="100"></a><br/> [![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go) ![Build Status](https://github.com/ewohltman/ephemeral-roles/workflows/build/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ewohltman/ephemeral-roles)](https://goreportcard.com/report/github.com/ewohltman/ephemeral-roles) [![Coverage Status](https://coveralls.io/repos/github/ewohltman/ephemeral-roles/badge.svg?branch=master)](https://coveralls.io/github/ewohltman/ephemeral-roles?branch=master) [![Dev Chat](https://img.shields.io/badge/discord-dev%20chat-blue)](https://discord.gg/yrxaJb5) |
| :------: |

### A Discord bot for managing ephemeral roles based upon voice channel member presence.

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

## What does `Ephemeral Roles` do?

After the `Ephemeral Roles` bot is invited to your Discord server, it
immediately starts to watch for changes to your voice channels.  When a member
joins a channel, `Ephemeral Roles` automatically assigns that member an
*ephemeral role* associated with the channel.  If the *ephemeral role* doesn't
exist yet, `Ephemeral Roles` will create it.

By having your members auto-sorted into *ephemeral roles* in your member list,
it's clear to see who are available for chatting and the channels they are in.
This is because `Ephemeral Roles` leverages the Discord feature that the member
list in servers will group together members by role right out of the box.

When a member changes or disconnects from voice channels, even across Discord
servers, `Ephemeral Roles` will account for the change and automatically
revoke/reissue *ephemeral roles* as appropriate.

----

## Example Usage

| Orange roles below are automatically managed by `Ephemeral Roles` |
| :------: |
| ![Ephemeral Roles action example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/action.gif) |
| ![Ephemeral Roles static example](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/static.png) |
| ![Ephemeral Roles example role list](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/roles.png) |

----

## Monitoring

A **[Prometheus](https://prometheus.io/)** and **[Grafana](https://grafana.com/)** instance have been set up to monitor `Ephemeral Roles` metrics.

| [grafana.ephemeral-roles.net](http://grafana.ephemeral-roles.net/d/OqANQqtiz/ephemeral-roles-metrics?orgId=1&refresh=5s) |
| :------: |
| <a href="http://grafana.ephemeral-roles.net/d/OqANQqtiz/ephemeral-roles-metrics?orgId=1&refresh=5s"><img src="https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/bot-metrics.png"></a> |

----

## Architecture

| Architectural Diagram |
| :------: |
| ![Architecture](https://raw.githubusercontent.com/ewohltman/ephemeral-roles/master/web/static/architecture.png) |
| Architectural diagram created with [draw.io](https://draw.io/) |

* `ephemeral-roles`:
  * Runs in a Kubernetes cluster as a StatefulSet of 10 Pods. Each Pod contains
    a running instance of the bot
  * A StatefulSet is used so that each Pod has a predictable name so that the
    bot instance can inform the Discord API which of the 10 Pods it is
  * The Discord API will assign a number of the total guilds (servers) to each
    of the bot instances to balance the load of managing the guild events
  * If any of the Pods stop running for whatever reason, the StatefulSet will
    automatically restart them
* `pod-bouncer` (https://github.com/ewohltman/pod-bouncer):
  * Runs in a Pod and is responsible for receiving alerts from
    `Prometheus`/`AlertManager` and to act upon them by automatically causing
    unhealthy Pods for `ephemeral-roles` to restart
* `ephemeral-roles-informer` (https://github.com/ewohltman/ephemeral-roles-informer):
  * Runs in a Pod and is responsible for collecting metrics from the
    `ephemeral-roles` instances to update search services such as
    [discord.bots.gg](https://discord.bots.gg/) and [top.gg](https://top.gg/)

----

## Contributing to the project

Contributions are very welcome! Please follow the guidelines below:

* Open an issue describing the bug or enhancement
* Fork the `develop` branch and make your changes
  * Try to match current naming conventions as closely as possible
  * Try to keep changes small and incremental with appropriate new unit tests
* Create a Pull Request with your changes against the `develop` branch

This project is equipped with a full
[CI](https://en.wikipedia.org/wiki/Continuous_integration)
/
[CD](https://en.wikipedia.org/wiki/Continuous_deployment) pipeline:
 
* Linting and unit tests will be automatically run with the PR, providing
feedback if any additional changes need to be made
* Merge to `master` will automatically deploy the changes live
