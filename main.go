package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/alexadhy/shortener/handlers"
	"github.com/alexadhy/shortener/internal/log"
	"github.com/alexadhy/shortener/internal/middlewares"
	"github.com/alexadhy/shortener/persist/redis"
)

const (
	defaultRedisAddr = "localhost:6379"
	defaultPort      = "8388"
	defaultHost      = "localhost"
)

func main() {
	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")
	domain := os.Getenv("APP_DOMAIN")
	redisAddr := os.Getenv("APP_REDIS_ADDRESSES")

	if redisAddr == "" {
		redisAddr = defaultRedisAddr
	}

	if host == "" {
		host = defaultHost
	}

	if port == "" {
		port = defaultPort
	}

	if domain == "" {
		domain = host
	}

	listenAddress := fmt.Sprintf("http://" + host + ":" + port)
	redisAddrs := strings.Split(redisAddr, " ")

	router := chi.NewRouter()
	//lmt := tollbooth.NewLimiter(3, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	router.Use(middleware.RequestID)
	router.Use(middlewares.LoggerMW())
	//router.Use(middlewares.LimitHandler(lmt))
	router.Use(middleware.Recoverer)

	store, err := redis.New(redisAddrs...)
	if err != nil {
		log.Fatalf("redis.New(): %v", err)
	}
	apiSrv := handlers.New(store, listenAddress, func(s string) bool {
		return true
	})

	router.Post("/", apiSrv.CreateShortLink)
	router.Get("/{id}", apiSrv.HandleRedirect)

	server := http.Server{Addr: defaultHost + ":" + defaultPort, Handler: router}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	log.Infof("Listening to HTTP requests at %s", listenAddress)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
