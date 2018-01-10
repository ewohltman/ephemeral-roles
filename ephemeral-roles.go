package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/ewohltman/discordgo"
	"github.com/ewohltman/ephemeral-roles/pkg/callbacks"
	"github.com/ewohltman/ephemeral-roles/pkg/discordBotsOrg"
	"github.com/ewohltman/ephemeral-roles/pkg/logging"
	"github.com/sirupsen/logrus"
)

// internalStateCache is a mutex protected cache of values
type internalStateCache struct {
	mu        sync.RWMutex
	guildList []*discordgo.Guild
	numGuilds int
}

// Server is the struct for the internal HTTP server
type Server struct {
	logger *logrus.Logger
	mux    *http.ServeMux
}

var log = logging.Instance()
var isc *internalStateCache

func newServer(options ...func(*Server)) *Server {
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

	// List the guilds our bot is a member of
	s.mux.HandleFunc(
		"/guilds",
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			isc.mu.RLock()
			defer isc.mu.RUnlock()

			buf := bytes.NewBuffer([]byte{})
			for _, guild := range isc.guildList {
				buf.Write([]byte(guild.Name + "\n"))
			}

			response := buf.Bytes()

			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(response)))

			_, err := w.Write(response)
			if err != nil {
				log.WithError(err).Errorf("Error writing /guilds HTTP response")
				return
			}
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

func monitorGuildsUpdate(dgBotSession *discordgo.Session, token string, botID string) {
	for true {
		isc.mu.RLock()
		checkNum := isc.numGuilds
		isc.mu.RUnlock()

		if len(dgBotSession.State.Guilds) > checkNum {
			log.WithField(
				"guild",
				dgBotSession.State.Guilds[len(dgBotSession.State.Guilds)-1].Name,
			).Infof(dgBotSession.State.User.Username + " joined new guild")

			guildsUpdate(dgBotSession, token, botID)
		} else {
			log.Infof(dgBotSession.State.User.Username + " removed from guild")

			guildsUpdate(dgBotSession, token, botID)
		}

		time.Sleep(time.Second * 5)
	}
}

func guildsUpdate(dgBotSession *discordgo.Session, token string, botID string) {
	isc.mu.Lock()
	defer isc.mu.Unlock()

	isc.guildList = dgBotSession.State.Guilds
	isc.numGuilds = len(isc.guildList)

	if token != "" && botID != "" {
		discordBotsOrg.Update(token, botID, isc.numGuilds)
	}
}

func main() {
	log.Debugf("Bot starting up")

	// Check for BOT_TOKEN, we need this to connect to Discord
	token, found := os.LookupEnv("BOT_TOKEN")
	if !found || token == "" {
		log.Fatalf("BOT_TOKEN not defined in environment variables")
	}
        
	// Check for string from slice, these are not needed now, but are needed in the callbacks
	for _, envVar := range []string{"BOT_NAME", "BOT_KEYWORD", "ROLE_PREFIX"} {
		_, found = os.LookupEnv(envVar)
		if !found {
			log.Fatalf("%s not defined in environment variables")
		}
	}

	// Check for PORT, we need this to for our HTTP server in our container
	port, found := os.LookupEnv("PORT")
	if !found || port == "" {
		port = "8080"
	}

	// Check for DISCORDBOTS_ORG_TOKEN and BOT_ID, we need these for optional discordbots.org integration
	discordBotsToken := ""
	botID := ""

	discordBotsToken, found = os.LookupEnv("DISCORDBOTS_ORG_TOKEN")
	if !found || discordBotsToken == "" {
		log.WithField("warn", "DISCORDBOTS_ORG_TOKEN not defined in environment variables").
			Warnf("Integration with discordbots.org integration disabled")
	} else {
		botID, found = os.LookupEnv("BOT_ID")
		if !found || botID == "" {
			log.WithField("warn", "BOT_ID not defined in environment variables").
				Warnf("Integration with discordbots.org integration disabled")
		}
	}

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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP)
	signal.Notify(stop, os.Interrupt)

	httpServer := &http.Server{
		Addr: ":" + port,
		Handler: newServer(func(s *Server) {
			s.logger = log
		},
		),
	}

	isc = &internalStateCache{}
	guildsUpdate(dgBotSession, discordBotsToken, botID)
	go monitorGuildsUpdate(dgBotSession, discordBotsToken, botID)

	log.Debugf("Starting internal HTTP server instance")
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.WithError(err).Errorf("Internal HTTP server error")
		}
	}()

	// Block until the OS signal
	<-stop

	log.Warnf("Caught graceful shutdown signal")

	// Cleanly close down the Discord session
	dgBotSession.Close()

	// Cleanly shutdown the HTTP server
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()

	httpServer.Shutdown(ctx)

	return
}
