package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/config"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/ewohltman/ephemeral-roles/pkg/monitor"
	"github.com/ewohltman/ephemeral-roles/pkg/server"
)

var (
	token string
	port  string
	botID string
	log   = logging.Instance()
)

func checkEnvironmentConfig() {
	var err error

	token, port, err = config.CheckRequired()
	if err != nil {
		log.Fatal(err)
	}

	_, _, err = config.CheckDiscordBotsOrg()
	if err != nil {
		log.Warn(err)
	}
}

func runHTTPServer() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP)
	signal.Notify(stop, os.Interrupt)

	httpServer := server.New(port)

	log.Debugf("Starting internal HTTP server instance")
	go func() {
		if serverError := httpServer.ListenAndServe(); serverError != nil {
			log.WithError(serverError).Warnf("Internal HTTP server")
		}
	}()

	// Block until the OS signal
	<-stop

	log.Warnf("Caught graceful shutdown signal")

	// Cleanly shutdown the HTTP server
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	err := httpServer.Shutdown(ctx)
	if err != nil {
		log.Error(err)
	}
}

func main() {
	log.Debugf("Bot starting up")

	checkEnvironmentConfig()

	// Create a new Discord session using the provided bot token
	dgBotSession, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	// Add event handlers
	dgBotSession.AddHandler(callbacks.Ready)            // Connection established with Discord
	dgBotSession.AddHandler(callbacks.MessageCreate)    // Chat messages with BOT_KEYWORD
	dgBotSession.AddHandler(callbacks.VoiceStateUpdate) // Updates to voice channel state

	// Open the websocket and begin listening
	err = dgBotSession.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}
	defer dgBotSession.Close()

	monitor.Start(dgBotSession)

	runHTTPServer()
}
