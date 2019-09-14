package ephemeral_roles

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Check for BOT_ID and DISCORDBOTS_ORG_TOKEN, we need these for optional discordbots.org integration
	botID, found := os.LookupEnv("BOT_ID")
	if !found || botID == "" {
		return nil, errors.New("integration with discordbots.org disabled: BOT_ID not defined in environment variables")
	}

	discordBotsOrgToken, found := os.LookupEnv("DISCORDBOTS_ORG_TOKEN")
	if !found || discordBotsOrgToken == "" {
		return nil, errors.New("integration with discordbots.org disabled: DISCORDBOTS_ORG_TOKEN not defined in environment variables")
	}

	return &optionalConfig{
		botID:               botID,
		discordBotsOrgToken: discordBotsOrgToken,
	}, nil
}

func main() {
	log := logging.New()

	log.Debugf("Bot starting up")

	required, err := checkRequired()
	if err != nil {
		log.WithError(err).Fatal("Missing required environment variable")
	}

	_, err = checkOptional()
	if err != nil {
		log.WithError(err).Warn("Missing required environment variable")
	}

	// Create a new Discord session using the provided bot token
	session, err := discordgo.New("Bot " + required.token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	// Open the websocket and begin listening
	err = session.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}
	defer session.Close()

	callbackMetrics := monitor.Start(
		&monitor.Config{
			Log:                 log,
			Session:             session,
			BotID:               "",
			DiscordBotsOrgToken: "",
		},
	)

	callbackConfig := &callbacks.Config{
		Log:                     log,
		BotName:                 os.Getenv("BOT_NAME"),
		BotKeyword:              os.Getenv("BOT_KEYWORD"),
		RolePrefix:              os.Getenv("ROLE_PREFIX"),
		ReadyCounter:            callbackMetrics.ReadyCounter,
		MessageCreateCounter:    callbackMetrics.MessageCreateCounter,
		VoiceStateUpdateCounter: callbackMetrics.VoiceStateUpdateCounter,
	}

	// Add event handlers
	session.AddHandler(callbackConfig.Ready)            // Connection established with Discord
	session.AddHandler(callbackConfig.MessageCreate)    // Chat messages with BOT_KEYWORD
	session.AddHandler(callbackConfig.VoiceStateUpdate) // Updates to voice channel state

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	signal.Notify(stop, os.Interrupt)

	httpServer := server.New(log, required.port)

	log.Debugf("Starting internal HTTP server instance")
	go func() {
		if serverError := httpServer.ListenAndServe(); serverError != nil {
			log.WithError(serverError).Error("Internal server error")
		}
	}()

	// Block until the OS signal
	<-stop

	log.Warnf("Caught graceful shutdown signal")

	// Cleanly shutdown the HTTP server
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.WithError(err).Error("Error shutting down server")
	}
}
