// Command driver launches ephemeral-roles' real HTTP server and callback
// handler against an in-memory mock Discord session, so the bot can be driven
// and observed without a live Discord connection or a BOT_TOKEN.
//
// The production binary (cmd/ephemeral-roles) cannot run headless: it calls
// session.Open() at startup and dies with "Authentication failed" against the
// live Discord gateway. This driver swaps in internal/pkg/mock.NewSession()
// (backed by github.com/ewohltman/discordgo-mock) which serves Discord REST
// calls from in-process state, then wires up the SAME production types:
//
//   - internal/pkg/http.NewServer          -> /, /guilds, /metrics, pprof
//   - internal/pkg/monitor.NewMetrics      -> ephemeral_roles_* gauges/counters
//   - internal/pkg/callbacks.Handler       -> VoiceStateUpdate core flow
//
// It lives under .claude/ so it is excluded from `go build ./...`,
// `go vet ./...`, and golangci-lint (the go toolchain ignores dot-dirs), but
// it can still import the module's internal packages because it sits inside
// the module tree. Run it explicitly by file path:
//
//	go run ./.claude/skills/run-ephemeral-roles/driver.go
//
// On start it fires one real VoiceStateUpdate (a member joining a voice
// channel), prints the guild's ephemeral roles before and after so you can see
// the role get created and assigned, then serves HTTP until SIGINT/SIGTERM so
// you can curl the endpoints.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/discordgo-mock/mockconstants"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/mock"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const (
	rolePrefix = "{eph}"
	roleColor  = 16753920
)

func main() {
	addr := flag.String("addr", "127.0.0.1:18099", "host:port for the HTTP server")
	serve := flag.Bool("serve", true, "keep the HTTP server running after the callback demo (SIGINT/SIGTERM to stop)")
	flag.Parse()

	log := logging.New(logging.OptionalLogLevel("info"))

	jaegerTracer, jaegerCloser, err := tracer.New("ephemeral-roles-driver")
	if err != nil {
		log.WithError(err).Fatal("Error creating Jaeger tracer")
	}
	defer func() { _ = jaegerCloser.Close() }()

	session, err := mock.NewSession()
	if err != nil {
		log.WithError(err).Fatal("Error creating mock session")
	}
	defer func() { _ = session.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Short interval so the guilds/members gauges populate within a second or
	// two of startup (they are only Set on each ticker tick, never eagerly).
	metrics := monitor.NewMetrics(&monitor.Config{
		Log:      log,
		Session:  session,
		Interval: time.Second,
	})
	metrics.Monitor(ctx)

	handler := &callbacks.Handler{
		Log:                     log,
		RolePrefix:              rolePrefix,
		RoleColor:               roleColor,
		JaegerTracer:            jaegerTracer,
		ReadyCounter:            metrics.ReadyCounter,
		VoiceStateUpdateCounter: metrics.VoiceStateUpdateCounter,
		OperationsGateway:       operations.NewGateway(session),
	}

	runCallbackDemo(log, session, handler)

	if !*serve {
		return
	}

	serveHTTP(log, session, *addr)
}

// runCallbackDemo fires one real VoiceStateUpdate (a member joining a voice
// channel) and prints the guild's role count and the member's assigned roles
// before and after, so the core role-creation/assignment flow is observable
// end-to-end.
//
// The member starts in guild "testGuild" already holding "{eph} testChannel".
// Joining "testChannel2" should: create a new ephemeral role for that channel,
// remove the member's stale "{eph} testChannel" role, and assign the new one.
// So expect the guild role count to grow by one and the member's role set to
// change. (The new role shows an empty name — see Gotchas in SKILL.md: the mock
// REST layer ignores the role-create request body.)
func runCallbackDemo(log logging.Interface, session *discordgo.Session, handler *callbacks.Handler) {
	guildID := mockconstants.TestGuild
	userID := mockconstants.TestUser
	channelID := mockconstants.TestChannel2

	log.Infof("=== VoiceStateUpdate demo: user %q joins voice channel %q in guild %q ===", userID, channelID, guildID)
	log.Infof("BEFORE: guild roles=%d, member roles=%s", guildRoleCount(session, guildID), memberRoles(session, guildID, userID))

	handler.VoiceStateUpdate(session, &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			GuildID:   guildID,
			UserID:    userID,
			ChannelID: channelID,
		},
	})

	log.Infof("AFTER:  guild roles=%d, member roles=%s", guildRoleCount(session, guildID), memberRoles(session, guildID, userID))
	log.Info("=== VoiceStateUpdate demo complete: guild gained a role and the member was assigned it => create+assign flow ran ===")
}

// guildRoleCount returns the number of roles currently in the guild's state.
func guildRoleCount(session *discordgo.Session, guildID string) int {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		return -1
	}

	return len(guild.Roles)
}

// memberRoles returns a member's assigned roles as "name(id-prefix)" pairs so a
// change in assignment is visible even when a role's name is unset.
func memberRoles(session *discordgo.Session, guildID, userID string) string {
	member, err := session.State.Member(guildID, userID)
	if err != nil {
		return fmt.Sprintf("<error: %s>", err)
	}

	parts := make([]string, 0, len(member.Roles))

	for _, roleID := range member.Roles {
		name := "<unnamed>"

		if role, roleErr := session.State.Role(guildID, roleID); roleErr == nil && role.Name != "" {
			name = role.Name
		}

		idPrefix := roleID
		if len(idPrefix) > 6 {
			idPrefix = idPrefix[:6]
		}

		parts = append(parts, fmt.Sprintf("%s(%s)", name, idPrefix))
	}

	sort.Strings(parts)

	return "[" + strings.Join(parts, ", ") + "]"
}

// serveHTTP starts the production HTTP server against the mock session and
// blocks until SIGINT/SIGTERM.
func serveHTTP(log logging.Interface, session *discordgo.Session, addr string) {
	host, port, ok := strings.Cut(addr, ":")
	if !ok {
		host, port = "127.0.0.1", addr
	}

	server := internalHTTP.NewServer(log, session, port)
	server.Addr = host + ":" + port

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Infof("HTTP server listening on http://%s  (endpoints: / , /guilds , /metrics , /debug/pprof/)", server.Addr)

		listenErr := server.ListenAndServe()
		if listenErr != nil && listenErr != http.ErrServerClosed {
			log.WithError(listenErr).Error("HTTP server error")

			stop <- syscall.SIGTERM
		}
	}()

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Error shutting down HTTP server")
	}

	log.Info("HTTP server stopped")
}
