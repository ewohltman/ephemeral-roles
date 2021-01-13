package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	stdLog "log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env"
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
)

type environmentVariables struct {
	BotToken             string `env:"BOT_TOKEN,required"`
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	LogTimezoneLocation  string `env:"LOG_TIMEZONE_LOCATION" envDefault:"UTC"`
	DiscordrusWebHookURL string `env:"DISCORDRUS_WEBHOOK_URL"`
	Port                 string `env:"PORT" envDefault:"8081"`
	BotName              string `env:"BOT_NAME" envDefault:"Ephemeral Roles"`
	BotKeyword           string `env:"BOT_KEYWORD" envDefault:"!eph"`
	RolePrefix           string `env:"ROLE_PREFIX" envDefault:"{eph}"`
	RoleColor            int    `env:"ROLE_COLOR_HEX2DEC" envDefault:"16753920"`
	InstanceName         string `env:"INSTANCE_NAME" envDefault:"ephemeral-roles-0"`
	ShardCount           int    `env:"SHARD_COUNT" envDefault:"1"`
	shardID              int
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
	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	callbackMetrics := monitor.NewMetrics(&monitor.Config{
		Log:      log,
		Session:  session,
		Interval: monitorInterval,
	})

	setupCallbackHandler(session,
		&callbacks.Handler{
			Log:                     log,
			BotName:                 envVars.BotName,
			BotKeyword:              envVars.BotKeyword,
			RolePrefix:              envVars.RolePrefix,
			RoleColor:               envVars.RoleColor,
			JaegerTracer:            jaegerTracer,
			ContextTimeout:          contextTimeout,
			ReadyCounter:            callbackMetrics.ReadyCounter,
			MessageCreateCounter:    callbackMetrics.MessageCreateCounter,
			VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
			OperationsGateway:       operations.NewGateway(session),
		},
	)

	err = session.Open()
	if err != nil {
		return nil, err
	}

	callbackMetrics.Monitor(ctx)

	return session, nil
}

func setupCallbackHandler(session *discordgo.Session, callbackConfig *callbacks.Handler) {
	session.AddHandler(callbackConfig.ChannelDelete)
	session.AddHandler(callbackConfig.MessageCreate)
	session.AddHandler(callbackConfig.Ready)
	session.AddHandler(callbackConfig.VoiceStateUpdate)
}

func startHTTPServer(log logging.Interface, session *discordgo.Session, port string) (httpServer *http.Server, stop chan os.Signal) {
	httpServer = internalHTTP.NewServer(log, session, port)
	stop = make(chan os.Signal, 1)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Error("HTTP server error")
				stop <- syscall.SIGTERM
			}
		}
	}()

	signal.Notify(stop, syscall.SIGTERM)

	return httpServer, stop
}

func closeComponent(log logging.Interface, component string, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.WithError(err).Errorf("Error closing %s", component)
	}
}

func main() {
	envVars := &environmentVariables{}

	err := env.Parse(envVars)
	if err != nil {
		stdLog.Fatalf("Error looking up environment variables: %s", err)
	}

	err = envVars.parseShardID()
	if err != nil {
		stdLog.Fatalf("Error parsing shard ID: %s", err)
	}

	log := logging.New(
		logging.OptionalShardID(envVars.shardID),
		logging.OptionalLogLevel(envVars.LogLevel),
		logging.OptionalTimezoneLocation(envVars.LogTimezoneLocation),
		logging.OptionalDiscordrus(envVars.DiscordrusWebHookURL),
	)

	log.Infof("%s starting up", envVars.BotName)

	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic! %v", r)
		}
	}()

	jaegerTracer, jaegerCloser, err := tracer.New(ephemeralRoles)
	if err != nil {
		log.WithError(err).Fatal("Error setting up Jaeger tracer")
	}

	defer closeComponent(log, "Jaeger tracer", jaegerCloser)

	client := internalHTTP.NewClient(internalHTTP.WrapTransport(
		internalHTTP.NewTransport(),
		internalHTTP.WrapTransportWithTracer(jaegerTracer, envVars.InstanceName),
	))

	monitorCtx, cancelMonitorCtx := context.WithCancel(context.Background())
	defer cancelMonitorCtx()

	session, err := startSession(monitorCtx, log, envVars, client, jaegerTracer)
	if err != nil {
		log.WithError(err).Fatal("Error starting Discord session")
	}

	defer closeComponent(log, "Discord session", session)

	httpServer, stop := startHTTPServer(log, session, envVars.Port)

	<-stop // Block until the OS signal

	shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelShutdownCtx()

	err = httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.WithError(err).Error("Error shutting down HTTP server gracefully")
	}
}
