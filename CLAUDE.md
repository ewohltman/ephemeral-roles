# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`ephemeral-roles` is a Discord bot (Go) that watches voice channel presence and auto-assigns/revokes
"ephemeral roles" (prefixed, e.g. `{eph} General`) matching the channel a member is currently in. It runs
as a StatefulSet of shards in Kubernetes; each pod is one shard of the same Discord application.

## Commands

All common tasks go through the `Makefile` (uses `go tool -modfile=tools/go.mod` to run pinned tool
versions from `tools/go.mod`, separate from the main module):

```
make tidy    # go mod tidy for both the root module and tools/ module
make fmt     # gofmt -s -w + goimports (local-prefix github.com/ewohltman/ephemeral-roles/), runs tidy first
make vet     # go vet ./...
make lint    # vet + golangci-lint run ./... (config: .golangci.yml, "all" linters minus an explicit disable list)
make test    # gotestsum with -race and coverage, excludes ./cmd/...; prints total coverage
make build   # CGO_ENABLED=0 trimpath build -> build/package/ephemeral-roles/ephemeral-roles
make image   # podman build of the Docker image (build/package/ephemeral-roles/Dockerfile)
```

Run a single test package/test the normal Go way, e.g.:
```
go test ./internal/pkg/callbacks/... -run TestHandler_VoiceStateUpdate -race
```

CI (`.github/workflows/pullRequest.yml`) runs `make tidy`, `make vet`, `golangci-lint` (golangci-lint-action v9, golangci-lint v2.12.2), `make test`,
`make build`, `make image` on every PR targeting non-master branches. `pullRequestMaster.yml` covers PRs into
master. Development happens on `develop` (the repo's default branch); contributions target `develop` per
CONTRIBUTING.md, and merges to `master` auto-deploy.

## Architecture

Entry point: `cmd/ephemeral-roles/ephemeral-roles.go`. On startup it: parses env vars (via `caarlos0/env`,
see `environmentVariables` struct), derives a shard ID from `INSTANCE_NAME` (expects a trailing
`-<N>` from the StatefulSet pod name), builds a `logging.Logger`, a Jaeger tracer, an HTTP client wrapped
with tracing middleware, opens the `discordgo.Session`, registers callback handlers, starts Prometheus
monitoring goroutines, and starts the HTTP server. Shutdown is on SIGINT/SIGTERM.

Package layout under `internal/pkg/`:

- **callbacks** — `Handler` holds bot config/dependencies and the Discord event callback methods
  (`Ready`, `VoiceStateUpdate`, `ChannelDelete`). `VoiceStateUpdate` is the core flow: look up
  guild/member/channel from `session.State`, resolve or create the `{prefix} <channel name>` role,
  remove any other ephemeral roles the member holds (by prefix match), then add the new one. Errors from
  this parse/lookup step are typed (`MemberNotFoundError`, `ChannelNotFoundError`, `RoleNotFoundError`,
  `InsufficientPermissionsError`, `MaxNumberOfRolesError`, `DeadlineExceededError` in `errors.go`), all
  implementing a shared `CallbackError` interface (`Is`/`Unwrap`/`InGuild`/`ForMember`/`InChannel`) so the
  handler can branch on error type and attach structured log fields without repeating guild/member/channel
  plumbing.
- **operations** — `Gateway` (backed by `golang.org/x/sync/singleflight`) centralizes and de-duplicates
  Discord API-mutating requests (currently `CreateRole`) so concurrent `VoiceStateUpdate` callbacks racing
  to create the same role collapse into a single API call. Also holds guild/role/member lookup and
  permission-check helpers used directly by callbacks (`LookupGuild`, `AddRoleToMember`,
  `RemoveRoleFromMember`, `BotHasChannelPermission`), plus classifiers for specific Discord REST errors
  (`IsForbiddenResponse`, `IsMaxGuildsResponse`, `IsDeadlineExceeded`, `ShouldLogDebug`).
- **monitor** — background goroutines (`Guilds`, `Members`) that periodically poll session state and
  update Prometheus gauges/counters (guild count, member count, Ready/VoiceStateUpdate event totals),
  namespaced `ephemeral_roles`.
- **http** — the bot's own HTTP server (`NewServer`): `/` health root, `/guilds` (JSON list of guilds
  sorted by member count), `/metrics` (Prometheus), and pprof endpoints. Also provides the outbound
  `http.Client`/`http.Transport` constructors (`NewClient`, `NewTransport`) for Discord API calls — the
  Jaeger tracing round tripper is layered on via `tracer.RoundTripper` in `cmd` — plus `ErrorLogger`, an
  `slog`-backed `*log.Logger` used for the server's `ErrorLog`.
- **logging** — wraps the standard library `log/slog`. `New` returns a `*Logger` (embedding
  `*slog.Logger`) built on a custom fan-out `slog.Handler` that writes to stdout and, when a webhook URL
  is configured, also to Discord via `github.com/Bufferoverflovv/slog-discord`. Supports runtime
  log-level updates (through a shared `slog.LevelVar`), a configurable timezone (a `ReplaceAttr` hook),
  and adapts `discordgo`'s own logger through `DiscordGoLogf`. A concrete `*slog.Logger` (not an
  interface) is passed around the codebase.
- **tracer** — Jaeger/OpenTracing setup and an `http.RoundTripper` middleware that wraps outbound calls
  in spans.
- **mock** — test doubles: a mirror `RoundTripper`, a discarding `*slog.Logger` (`NewLogger`), and (`session.go`)
  a pre-populated `*discordgo.Session` builder built on the separate `github.com/ewohltman/discordgo-mock`
  module (guild/role/member/channel/state mocks), used across `_test.go` files instead of hitting the
  real Discord API.

`tools/` is an independent Go module (own `go.mod`) that pins developer tooling (`goimports`,
`golangci-lint`, `gotestsum`) via `go tool`, keeping those versions out of the main module's dependency
graph.

## Conventions

- Errors are wrapped with `%w` and typed where callers need to branch on them (see `callbacks/errors.go`
  for the pattern: struct holds context like `Guild`/`Member`/`Channel` plus `Err`, implements
  `Error()`/`Is()`/`Unwrap()`).
- A concrete `*slog.Logger` is threaded through constructors rather than a global logger.
- golangci-lint runs with `default: all` linters and an explicit `disable` list in `.golangci.yml` — when
  adding code, prefer satisfying the stricter defaults (e.g. `wsl_v5` whitespace/cuddling rules, `cyclop`/
  `gocyclo` complexity limits of 15, `funlen` of 100 lines/50 statements, `lll` at 140 chars) rather than
  adding new exclusions. Test files (`_test.go`) already have a relaxed exclusion set (see the file).
