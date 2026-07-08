// Package main is the main package of the project.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v11"
	"github.com/opentracing/opentracing-go"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/operations"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const (
	ephemeralRoles  = "ephemeral-roles"
	contextTimeout  = 5 * time.Minute
	monitorInterval = 10 * time.Second

	intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentGuildMembers | discordgo.IntentGuildPresences
)

type environmentVariables struct {
	BotToken            string `env:"BOT_TOKEN,required"`
	DiscordWebhookURL   string `env:"DISCORD_WEBHOOK_URL"`
	LogLevel            string `env:"LOG_LEVEL"             envDefault:"info"`
	LogTimezoneLocation string `env:"LOG_TIMEZONE_LOCATION" envDefault:"UTC"`
	Port                string `env:"PORT"                  envDefault:"8081"`
	BotName             string `env:"BOT_NAME"              envDefault:"Ephemeral Roles"`
	RolePrefix          string `env:"ROLE_PREFIX"           envDefault:"{eph}"`
	RoleColor           int    `env:"ROLE_COLOR_HEX2DEC"    envDefault:"16753920"`
	InstanceName        string `env:"INSTANCE_NAME"         envDefault:"ephemeral-roles-0"`
	ShardCount          int    `env:"SHARD_COUNT"           envDefault:"1"`
	shardID             int
}

func (envVars *environmentVariables) parseShardID() error {
	shardIDRegEx := regexp.MustCompile(`-\d.*$`)

	shardIDString := shardIDRegEx.FindString(envVars.InstanceName)
	shardIDString = strings.TrimPrefix(shardIDString, "-")

	shardID, err := strconv.Atoi(shardIDString)
	if err != nil {
		return fmt.Errorf("error parsing shard ID: %w", err)
	}

	envVars.shardID = shardID

	return nil
}

func run() error {
	ctx, cancelCtx := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancelCtx()

	ev := &environmentVariables{}

	if err := env.Parse(ev); err != nil {
		return fmt.Errorf("error parsing environment variables: %w", err)
	}

	if err := ev.parseShardID(); err != nil {
		return fmt.Errorf("error parsing shard ID: %w", err)
	}

	log := logging.New(
		logging.OptionalShardID(ev.shardID),
		logging.OptionalLogLevel(ev.LogLevel),
		logging.OptionalTimezoneLocation(ev.LogTimezoneLocation),
		logging.OptionalDiscordrus(ev.DiscordWebhookURL),
	)

	log.Infof("%s starting up", ev.BotName)

	jaegerTracer, jaegerCloser, err := tracer.New(ephemeralRoles)
	if err != nil {
		return fmt.Errorf("error creating Jaeger tracer: %w", err)
	}

	defer func() { _ = jaegerCloser.Close() }()

	client := internalHTTP.NewClient(tracer.RoundTripper(
		jaegerTracer,
		ev.InstanceName,
		internalHTTP.NewTransport(),
	))

	session, err := startSession(ctx, log, ev, client, jaegerTracer)
	if err != nil {
		return fmt.Errorf("error starting Discord session: %w", err)
	}

	defer func() { _ = session.Close() }()

	return runServer(ctx, log, session, ev.Port)
}

func startSession(
	ctx context.Context,
	log logging.Interface,
	envVars *environmentVariables,
	client *http.Client,
	jaegerTracer opentracing.Tracer,
) (*discordgo.Session, error) {
	discordgo.Logger = log.DiscordGoLogf

	session, err := discordgo.New("Bot " + envVars.BotToken)
	if err != nil {
		return nil, err
	}

	session.Client = client
	session.ShardID = envVars.shardID
	session.ShardCount = envVars.ShardCount
	session.LogLevel = discordgo.LogInformational
	session.State.TrackEmojis = false
	session.State.TrackPresences = false
	session.Identify.Intents = discordgo.MakeIntent(intents)

	callbackMetrics := monitor.NewMetrics(&monitor.Config{
		Log:      log,
		Session:  session,
		Interval: monitorInterval,
	})

	addCallbackHandlers(session,
		&callbacks.Handler{
			Log:                     log,
			RolePrefix:              envVars.RolePrefix,
			RoleColor:               envVars.RoleColor,
			JaegerTracer:            jaegerTracer,
			ReadyCounter:            callbackMetrics.ReadyCounter,
			VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
			OperationsGateway:       operations.NewGateway(session),
		},
	)

	if err := session.Open(); err != nil {
		return nil, err
	}

	callbackMetrics.Monitor(ctx)

	return session, nil
}

func addCallbackHandlers(session *discordgo.Session, callbackConfig *callbacks.Handler) {
	session.AddHandler(callbackConfig.Ready)
	session.AddHandler(callbackConfig.VoiceStateUpdate)
	session.AddHandler(callbackConfig.ChannelDelete)
}

func runServer(
	ctx context.Context,
	log logging.Interface,
	session *discordgo.Session,
	port string,
) error {
	httpServer := internalHTTP.NewServer(log, session, port)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Error("HTTP server error")
			}
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), contextTimeout)
	defer cancel()

	return httpServer.Shutdown(shutdownCtx)
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "fatal error: %s", err)
	}
}
