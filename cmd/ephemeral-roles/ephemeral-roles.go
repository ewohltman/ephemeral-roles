// Package main is the main package of the project.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
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

	intents = gateway.IntentGuilds |
		gateway.IntentGuildVoiceStates |
		gateway.IntentGuildMembers |
		gateway.IntentGuildPresences
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
		logging.OptionalDiscordWebhook(ev.DiscordWebhookURL),
	)

	log.Info("starting up", "bot", ev.BotName)

	jaegerTracer, jaegerCloser, err := tracer.New(ephemeralRoles)
	if err != nil {
		return fmt.Errorf("error creating Jaeger tracer: %w", err)
	}

	defer func() { _ = jaegerCloser.Close() }()

	httpClient := internalHTTP.NewClient(tracer.RoundTripper(
		jaegerTracer,
		ev.InstanceName,
		internalHTTP.NewTransport(),
	))

	client, err := startSession(ctx, log.Logger, ev, httpClient, jaegerTracer)
	if err != nil {
		return fmt.Errorf("error starting Discord session: %w", err)
	}

	defer client.Close(context.WithoutCancel(ctx))

	return runServer(ctx, log.Logger, client, ev.Port)
}

func startSession(
	ctx context.Context,
	log *slog.Logger,
	envVars *environmentVariables,
	httpClient *http.Client,
	jaegerTracer opentracing.Tracer,
) (*bot.Client, error) {
	client, err := disgo.New(envVars.BotToken,
		bot.WithLogger(log),
		bot.WithShardManagerConfigOpts(
			sharding.WithShardIDs(envVars.shardID),
			sharding.WithShardCount(envVars.ShardCount),
			sharding.WithAutoScaling(false),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(intents),
			),
		),
		bot.WithCacheConfigOpts(
			cache.WithCaches(
				cache.FlagGuilds,
				cache.FlagChannels,
				cache.FlagRoles,
				cache.FlagVoiceStates,
				cache.FlagMembers,
			),
		),
		bot.WithRestClientConfigOpts(
			rest.WithHTTPClient(httpClient),
		),
	)
	if err != nil {
		return nil, err
	}

	callbackMetrics := monitor.NewMetrics(&monitor.Config{
		Log:      log,
		Client:   client,
		Interval: monitorInterval,
	})

	addCallbackHandlers(client,
		&callbacks.Handler{
			Log:                     log,
			RolePrefix:              envVars.RolePrefix,
			RoleColor:               envVars.RoleColor,
			JaegerTracer:            jaegerTracer,
			ReadyCounter:            callbackMetrics.ReadyCounter,
			VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
			OperationsGateway:       operations.NewGateway(client),
		},
	)

	if err := client.OpenShardManager(ctx); err != nil {
		return nil, err
	}

	callbackMetrics.Monitor(ctx)

	return client, nil
}

func addCallbackHandlers(client *bot.Client, callbackConfig *callbacks.Handler) {
	client.AddEventListeners(
		bot.NewListenerFunc(callbackConfig.Ready),
		bot.NewListenerFunc(callbackConfig.VoiceStateUpdate),
		bot.NewListenerFunc(callbackConfig.ChannelDelete),
	)
}

func runServer(
	ctx context.Context,
	log *slog.Logger,
	client *bot.Client,
	port string,
) error {
	httpServer := internalHTTP.NewServer(log, client, port)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Error("HTTP server error", "error", err)
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
