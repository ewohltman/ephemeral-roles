package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/pkg/server"
)

type requiredConfig struct {
	token string
	port  string
}

type optionalConfig struct {
	botID               string
	discordBotsOrgToken string
}

func checkRequired() (*requiredConfig, error) {
	// Check for BOT_TOKEN, we need this to connect to Discord
	token, found := os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		return nil, errors.New("BOT_TOKEN not defined in environment variables")
	}

	// Check for PORT, we need this to for our HTTP server in our container
	port, found := os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}

	// Check for strings from slice, these are not needed now, but are needed in the callbacks
	for _, envVar := range []string{"BOT_NAME", "BOT_KEYWORD", "ROLE_PREFIX"} {
		value, found := os.LookupEnv(envVar)
		if !found || value == "" {
			return nil, errors.New("%s not defined in environment variables" + envVar)
		}
	}

	return &requiredConfig{
		token: token,
		port:  port,
	}, nil
}

func checkOptional() (*optionalConfig, error) {
	const integrationDisabled = "integration with discordbots.org disabled"

	// Check for BOT_ID and DISCORDBOTS_ORG_TOKEN, we need these for optional discordbots.org integration
	botID, found := os.LookupEnv("BOT_ID")
	if !found || botID == "" {
		return nil, errors.New(integrationDisabled + ": BOT_ID not defined in environment variables")
	}

	discordBotsOrgToken, found := os.LookupEnv("DISCORDBOTS_ORG_TOKEN")
	if !found || discordBotsOrgToken == "" {
		return nil, errors.New(integrationDisabled + ": DISCORDBOTS_ORG_TOKEN not defined in environment variables")
	}

	return &optionalConfig{
		botID:               botID,
		discordBotsOrgToken: discordBotsOrgToken,
	}, nil
}

func main() {
	log := logging.New()

	log.Info("Ephemeral Roles starting up")

	required, err := checkRequired()
	if err != nil {
		log.WithError(err).Fatal("Missing required environment variables")
	}

	_, err = checkOptional()
	if err != nil {
		log.WithError(err).Warn("Missing optional environment variables")
	}

	session, err := startSession(log, required.token)
	if err != nil {
		log.WithError(err).Fatal("Error starting Discord session")
	}

	defer func() {
		err := session.Close()
		if err != nil {
			log.WithError(err).Error("Error closing Discord session")
		}
	}()

	httpServer, stop := startHTTPServer(log, required)

	<-stop // Block until the OS signal

	log.Warnf("Caught graceful shutdown signal")

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("Error shutting down server")
	}
}

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

func startHTTPServer(log *logrus.Logger, required *requiredConfig) (*http.Server, chan os.Signal) {
	httpServer := server.New(log, required.port)

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
