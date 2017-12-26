package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ewohltman/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

var log = logging.Instance()

type Server struct {
	logger *logrus.Logger
	mux    *http.ServeMux
}

func NewServer(options ...func(*Server)) *Server {
	s := &Server{
		logger: log,
		mux:    http.NewServeMux(),
	}

	for _, f := range options {
		f(s)
	}

	// Do something special with /status later?
	s.mux.HandleFunc(
		"/status",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	s.mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
		},
	)

	return s
}

// (s *Server) ServeHTTP satisfies the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {
	// Check for BOT_TOKEN, we need this to connect to Discord
	token, found := os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}

	// Check for BOT_NAME, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("BOT_NAME")
	if !found {
		log.Fatalf("BOT_NAME not defined in environment variables")
	}

	// Check for BOT_KEYWORD, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("BOT_KEYWORD")
	if !found {
		log.Fatalf("BOT_KEYWORD not defined in environment variables")
	}

	// Check for ROLE_PREFIX, we don't need it now but it's required in the callbacks
	_, found = os.LookupEnv("ROLE_PREFIX")
	if !found {
		log.Fatalf("ROLE_PREFIX not defined in environment variables")
	}

	// Check for PORT, we need this to for our HTTP server in our container
	port, found := os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}

	// Create a new Discord session using the provided bot token
	dgBot, err := discordgo.New("Bot " + token)
	if err != nil {
		log.WithError(err).Fatalf("Error creating Discord session")
	}

	// Add event handlers
	dgBot.AddHandler(callbacks.Ready)            // Connection established with Discord
	dgBot.AddHandler(callbacks.MessageCreate)    // Chat messages with BOT_KEYWORD
	dgBot.AddHandler(callbacks.VoiceStateUpdate) // Updates to voice channel state

	// Open the websocket and begin listening
	err = dgBot.Open()
	if err != nil {
		log.WithError(err).Fatalf("Error opening Discord session")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	httpServer := &http.Server{
		Addr: ":" + port,
		Handler: NewServer(func(s *Server) {
			s.logger = log
		},
		),
	}

	go func() {
		log.Infof("Starting internal HTTP server instance")

		if err := httpServer.ListenAndServe(); err != nil {
			log.WithError(err).Errorf("Internal HTTP server error")
		}
	}()

	// Block until the OS signal
	<-stop

	log.Infof("Caught graceful shutdown signal")

	// Cleanly close down the Discord session
	dgBot.Close()

	// Cleanly shutdown the HTTP server
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	httpServer.Shutdown(ctx)

	return
}
