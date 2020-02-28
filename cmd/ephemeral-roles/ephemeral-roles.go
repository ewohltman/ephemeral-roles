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

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/environment"
	internalHTTP "github.com/ewohltman/ephemeral-roles/internal/pkg/http"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/tracer"
)

const (
	monitorInterval = 1 * time.Minute
	contextTimeout  = 20 * time.Second
)

func newLogger(variables *environment.Variables) *logging.Logger {
	return logging.New(variables.LogLevel, variables.LogTimezoneLocation, variables.DiscordrusWebHookURL)
}

func startSession(log logging.Interface, variables *environment.Variables, client *http.Client) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + variables.BotToken)
	if err != nil {
		return nil, err
	}

	session.Client = client
	session.ShardID = variables.ShardID
	session.ShardCount = variables.ShardCount

	monitorConfig := &monitor.Config{
		Log:                 log,
		Session:             session,
		HTTPClient:          session.Client,
		DiscordBotsOrgBotID: variables.DiscordBotsOrgBotID,
		DiscordBotsOrgToken: variables.DiscordBotsOrgToken,
		Interval:            monitorInterval,
	}

	setupCallbacks(monitorConfig, variables)

	err = session.Open()
	if err != nil {
		return nil, err
	}

	monitor.Start(monitorConfig)

	return session, nil
}

func setupCallbacks(monitorConfig *monitor.Config, variables *environment.Variables) {
	callbackMetrics := monitor.Metrics(monitorConfig)

	callbackConfig := &callbacks.Config{
		Log:                     monitorConfig.Log,
		BotName:                 variables.BotName,
		BotKeyword:              variables.BotKeyword,
		RolePrefix:              variables.RolePrefix,
		RoleColor:               variables.RoleColor,
		ReadyCounter:            callbackMetrics.ReadyCounter,
		MessageCreateCounter:    callbackMetrics.MessageCreateCounter,
		VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
	}

	monitorConfig.Session.AddHandler(callbackConfig.Ready)            // Connection established with Discord
	monitorConfig.Session.AddHandler(callbackConfig.MessageCreate)    // Chat messages with BOT_KEYWORD
	monitorConfig.Session.AddHandler(callbackConfig.VoiceStateUpdate) // Updates to voice channel state
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

	log.WithField("shardID", variables.ShardID).Infof("%s starting up", variables.BotName)

	jaegerTracer, jaegerCloser, err := tracer.New(log, variables.InstanceName)
	if err != nil {
		log.WithError(err).Fatal("Error setting up Jaeger tracer")
	}

	defer closeComponent(log, "Jaeger tracer", jaegerCloser)

	parentSpan := tracer.NewSpan(jaegerTracer, nil, variables.InstanceName)
	defer parentSpan.Finish()

	client := internalHTTP.NewClient(nil, jaegerTracer, parentSpan.Context())

	session, err := startSession(log, variables, client)
	if err != nil {
		log.WithError(err).Fatal("Error starting Discord session")
	}

	defer closeComponent(log, "Discord session", session)

	httpServer, stop := startHTTPServer(log, session, variables.Port)

	<-stop // Block until the OS signal

	ctx, cancelCtx := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelCtx()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("Error shutting down HTTP server gracefully")
	}
}
