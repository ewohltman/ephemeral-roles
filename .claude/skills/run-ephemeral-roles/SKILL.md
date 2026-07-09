---
name: run-ephemeral-roles
description: Build, run, and drive the ephemeral-roles Discord bot. Use when asked to start ephemeral-roles, run it, build it, exercise its VoiceStateUpdate/role flow, hit its HTTP/metrics endpoints, or run its tests.
---

`ephemeral-roles` is a Discord bot (Go, built on `github.com/disgoorg/disgo`)
that assigns/revokes voice-channel roles. The production binary **cannot run
headless** — it opens the disgo Discord gateway at startup and dies without a
real `BOT_TOKEN` + live Discord connection. So you drive it through
`.claude/skills/run-ephemeral-roles/driver.go`, which wires the **real**
`callbacks.Handler`, `monitor.Metrics`, and `http.Server` to the in-repo mock
Discord client (`internal/pkg/mock`: a disgo `*bot.Client` with a pre-populated
cache and a fake `rest.Rest`) — no token, no network. It fires one real
`VoiceStateUpdate` on startup and then serves the bot's HTTP endpoints so you
can `curl` them.

All paths below are relative to the repo root.

## Prerequisites

No `apt-get` packages needed beyond the Go toolchain and `curl` (both already
present in this container). The module requires **Go 1.26+** (`go version` →
`go1.26.4` here). `make build`/`make image` also reference `podman`, but the
driver does not need it.

## Build

The driver lives under `.claude/`, which the Go toolchain ignores for wildcard
patterns (`go build ./...`, `go vet ./...`, `golangci-lint run ./...` all skip
it), so it never pollutes CI. Build it by **explicit path**:

```bash
go build -o /tmp/er-driver ./.claude/skills/run-ephemeral-roles/
```

(Optional) the production binary still builds, just don't expect it to run
headless:

```bash
make build   # -> build/package/ephemeral-roles/ephemeral-roles
```

## Run (agent path)

Launch the driver in the background, wait for it to listen, then `curl` the
endpoints. **Capture the PID and `wait` on it after `kill`** (see Gotchas):

```bash
/tmp/er-driver -addr 127.0.0.1:18099 >/tmp/er-driver.log 2>&1 &
DRIVER_PID=$!
for i in $(seq 1 40); do curl -sf -o /dev/null http://127.0.0.1:18099/ && break; sleep 0.25; done
sleep 1.5   # let the guilds/members gauges tick at least once
curl -sS -o /dev/null -w "HTTP %{http_code}\n" http://127.0.0.1:18099/
curl -sS http://127.0.0.1:18099/guilds
curl -sS http://127.0.0.1:18099/metrics | grep -E '^ephemeral_roles_'
kill -TERM "$DRIVER_PID"
wait "$DRIVER_PID" 2>/dev/null   # required: reap the child before this shell exits
```

The startup log (`/tmp/er-driver.log`) shows the core role flow — a member
joining a voice channel triggers create + assign:

```
time=... level=INFO msg="BEFORE: guild roles=3, member roles=[testRole(1004), {eph} testChannel(1005)]"
time=... level=INFO msg="AFTER:  guild roles=4, member roles=[testRole(1004), {eph} testChannel2(1007)]"
```

The member starts holding `{eph} testChannel`; joining `testChannel2` creates a
new `{eph} testChannel2` role (role count 3→4), removes the stale
`{eph} testChannel`, and assigns the new one.

Verified endpoint output:

- `GET /` → `HTTP 200`, empty body (health root).
- `GET /guilds` → JSON, guilds sorted by member count desc (`testGuildLarge`
  3002, then `testGuild` 2).
- `GET /metrics` → Prometheus text; the bot's own gauges/counters are the
  `ephemeral_roles_*` lines:
  ```
  ephemeral_roles_guilds 2
  ephemeral_roles_members 3004
  ephemeral_roles_ready_events_total 0
  ephemeral_roles_voice_state_update_events_total 1
  ```
  `voice_state_update_events_total 1` confirms the callback fired.
- Also served: `/debug/pprof/`.

Flags: `-addr host:port` (default `127.0.0.1:18099`), `-serve=false` to run only
the callback demo and exit (quick smoke, no server):

```bash
/tmp/er-driver -serve=false   # fires one VoiceStateUpdate, prints BEFORE/AFTER, exits 0
```

## Run (human path)

The real bot (`make build` then run) needs a valid Discord bot token and reaches
out to Discord's gateway — it exits immediately in this container:

```bash
BOT_TOKEN=fake ./build/package/ephemeral-roles/ephemeral-roles
# -> fatal error: error starting Discord session: ... (disgo rejects the fake token / gateway connect fails)
```

Not usable headless; use the driver instead.

## Test

```bash
make test    # gotestsum, -race, coverage; excludes ./cmd/...
```

Expected: 6 packages under `internal/pkg/...` pass, ~80% total coverage, rc 0.
Run one package directly, e.g.:

```bash
go test ./internal/pkg/callbacks/... -run TestHandler_VoiceStateUpdate -race
```

## Gotchas

- **`go run ./.claude/.../driver.go` leaks the server.** `go run` spawns the
  compiled binary as a child and does not forward SIGTERM to it — killing the
  `go run` PID leaves the server bound to the port. Always `go build -o
  /tmp/er-driver` and run that single binary.
- **Reap the driver or the shell gets SIGSTKFLT.** If you `kill` the background
  driver but let the shell exit before the child is fully gone, the sandbox
  watchdog kills the shell (exit 144 = 128+16, no output). Always follow
  `kill -TERM "$DRIVER_PID"` with `wait "$DRIVER_PID"`.
- **Don't `pkill -f er-driver`.** The pattern matches the running shell's own
  command line and kills it mid-script. Use the captured `$DRIVER_PID`.
- **`guilds`/`members` gauges are 0 for the first second.** `monitor` only
  `Set`s them on each ticker tick, never eagerly. The driver uses a 1s interval;
  the `sleep 1.5` above lets them populate before you scrape `/metrics`.

## Troubleshooting

- **`/metrics` shows `ephemeral_roles_guilds 0`**: you scraped before the first
  monitor tick. Wait ~1.5s after startup and retry.
- **`bind: address already in use`**: a previous driver is still bound. Find it
  with `ss -ltnp | grep 18099` and `kill` that PID, or launch with a different
  `-addr`.
