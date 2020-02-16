package main

import (
	"context"
	"errors"
	stdLog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/ewohltman/ephemeral-roles/internal/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/environment"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/internal/pkg/server"
)

const (
	monitorInterval = 1 * time.Minute
	contextTimeout  = 5 * time.Second
)

func startSession(log logging.Interface, variables *environment.Variables) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + variables.BotToken)
	if err != nil {
		return nil, err
	}

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

func startHTTPServer(log logging.Interface, required *environment.Variables) (httpServer *http.Server, stop chan os.Signal) {
	httpServer = server.New(log, required.Port)
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

func main() {
	variables, err := environment.Lookup()
	if err != nil {
		stdLog.Fatalf("Error looking up environment variables: %s", err)
	}

	log := logging.New(variables)

	log.Infof("%s starting up", variables.BotName)

	session, err := startSession(log, variables)
	if err != nil {
		log.WithError(err).Fatal("Error starting Discord session")
	}

	defer func() {
		closeErr := session.Close()
		if closeErr != nil {
			log.WithError(closeErr).Error("Error closing Discord session")
		}
	}()

	httpServer, stop := startHTTPServer(log, variables)

	<-stop // Block until the OS signal

	ctx, cancelFunc := context.WithTimeout(context.Background(), contextTimeout)
	defer cancelFunc()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("Error shutting down HTTP server gracefully")
	}
}
