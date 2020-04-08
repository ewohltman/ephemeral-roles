package main

import (
	"context"
	"errors"
	"io"
	stdLog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/opentracing/opentracing-go"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/environment"
	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const (
	ephemeralRoles  = "ephemeral-roles"
	monitorInterval = 1 * time.Minute
	contextTimeout  = 30 * time.Second
)

func newLogger(variables *environment.Variables) *logging.Logger {
	return logging.New(
		variables.ShardID,
		variables.LogLevel,
		variables.LogTimezoneLocation,
		variables.DiscordrusWebHookURL,
	)
}

func startSession(
	ctx context.Context,
	log logging.Interface,
	variables *environment.Variables,
	client *http.Client,
	jaegerTracer opentracing.Tracer,
) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + variables.BotToken)
	if err != nil {
		return nil, err
	}

	monitorConfig := configureSession(log, session, client, jaegerTracer, variables)

	err = session.Open()
	if err != nil {
		return nil, err
	}

	monitor.Start(ctx, monitorConfig)

	return session, nil
}

func configureSession(
	log logging.Interface,
	session *discordgo.Session,
	client *http.Client,
	jaegerTracer opentracing.Tracer,
	variables *environment.Variables,
) *monitor.Config {
	monitorConfig := &monitor.Config{
		Log:      log,
		Session:  session,
		Interval: monitorInterval,
	}

	callbackMetrics := monitor.Metrics(monitorConfig)

	callbackConfig := &callbacks.Config{
		Log:                     log,
		BotName:                 variables.BotName,
		BotKeyword:              variables.BotKeyword,
		RolePrefix:              variables.RolePrefix,
		RoleColor:               variables.RoleColor,
		JaegerTracer:            jaegerTracer,
		ContextTimeout:          contextTimeout,
		ReadyCounter:            callbackMetrics.ReadyCounter,
		MessageCreateCounter:    callbackMetrics.MessageCreateCounter,
		VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
	}

	session.Client = client
	session.ShardID = variables.ShardID
	session.ShardCount = variables.ShardCount

	setupCallbacks(session, callbackConfig)

	return monitorConfig
}

func setupCallbacks(session *discordgo.Session, callbackConfig *callbacks.Config) {
	session.AddHandler(callbackConfig.Ready)            // Connection established with Discord
	session.AddHandler(callbackConfig.MessageCreate)    // Chat messages with BOT_KEYWORD
	session.AddHandler(callbackConfig.VoiceStateUpdate) // Updates to voice channel state
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
	variables, err := environment.Lookup()
	if err != nil {
		stdLog.Fatalf("Error looking up environment variables: %s", err)
	}

	log := newLogger(variables)

	log.Infof("%s starting up", variables.BotName)

	jaegerTracer, jaegerCloser, err := tracer.New(ephemeralRoles)
	if err != nil {
		log.WithError(err).Fatal("Error setting up Jaeger tracer")
	}

	defer closeComponent(log, "Jaeger tracer", jaegerCloser)

	client := internalHTTP.NewClient(nil, jaegerTracer, variables.InstanceName)

	monitorCtx, cancelMonitorCtx := context.WithCancel(context.Background())
	defer cancelMonitorCtx()

	session, err := startSession(monitorCtx, log, variables, client, jaegerTracer)
	if err != nil {
		log.WithError(err).Fatal("Error starting Discord session")
	}

	defer closeComponent(log, "Discord session", session)

	httpServer, stop := startHTTPServer(log, session, variables.Port)

	<-stop // Block until the OS signal

	shutdownCtx, cancelShutdownCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelShutdownCtx()

	err = httpServer.Shutdown(shutdownCtx)
	if err != nil {
		log.WithError(err).Error("Error shutting down HTTP server gracefully")
	}
}
