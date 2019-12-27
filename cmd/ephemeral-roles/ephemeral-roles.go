package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ewohltman/ephemeral-roles/pkg/environment"

	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/pkg/server"
)

func startSession(log *logrus.Logger, token string) (*discordgo.Session, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	monitorConfig := &monitor.Config{
		Log:                 log,
		Session:             session,
		HTTPClient:          session.Client,
		BotID:               "",
		DiscordBotsOrgToken: "",
		Interval:            1 * time.Minute,
	}

	setupCallbacks(monitorConfig)

	err = session.Open()
	if err != nil {
		return nil, err
	}

	monitor.Start(monitorConfig)

	return session, nil
}

func setupCallbacks(monitorConfig *monitor.Config) {
	callbackMetrics := monitor.Metrics(monitorConfig)

	callbackConfig := &callbacks.Config{
		Log:                     monitorConfig.Log,
		BotName:                 os.Getenv("BOT_NAME"),
		BotKeyword:              os.Getenv("BOT_KEYWORD"),
		RolePrefix:              os.Getenv("ROLE_PREFIX"),
		ReadyCounter:            callbackMetrics.ReadyCounter,
		MessageCreateCounter:    callbackMetrics.MessageCreateCounter,
		VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
	}

	monitorConfig.Session.AddHandler(callbackConfig.Ready)            // Connection established with Discord
	monitorConfig.Session.AddHandler(callbackConfig.MessageCreate)    // Chat messages with BOT_KEYWORD
	monitorConfig.Session.AddHandler(callbackConfig.VoiceStateUpdate) // Updates to voice channel state
}

func startHTTPServer(log *logrus.Logger, required *environment.RequiredVariables) (*http.Server, chan os.Signal) {
	httpServer := server.New(log, required.Port)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err.Error() != http.ErrServerClosed.Error() {
				log.WithError(err).Error("HTTP server error")
			}
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	signal.Notify(stop, os.Interrupt)

	return httpServer, stop
}

func main() {
	log := logging.New()

	log.Info("Ephemeral Roles starting up")

	requiredVariables, err := environment.CheckRequiredVariables()
	if err != nil {
		log.WithError(err).Fatal("Missing required environment variables")
	}

	_, err = environment.CheckOptionalVariables()
	if err != nil {
		log.WithError(err).Warn("Missing optional environment variables")
	}

	session, err := startSession(log, requiredVariables.BotToken)
	if err != nil {
		log.WithError(err).Fatal("Error starting Discord session")
	}

	defer func() {
		err := session.Close()
		if err != nil {
			log.WithError(err).Error("Error closing Discord session")
		}
	}()

	httpServer, stop := startHTTPServer(log, requiredVariables)

	<-stop // Block until the OS signal

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("Error shutting down server")
	}
}
