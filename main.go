package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/alexadhy/shortener/config"
	"github.com/alexadhy/shortener/handlers"
	"github.com/alexadhy/shortener/internal/log"
	"github.com/alexadhy/shortener/internal/middlewares"
	"github.com/alexadhy/shortener/persist/badger"
)

func main() {
	opts := config.New(func() config.Options {
		return config.Options{}
	})

	router := chi.NewRouter()
	lmt := tollbooth.NewLimiter(3, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	router.Use(middleware.RequestID)
	router.Use(middlewares.LoggerMW())
	router.Use(middlewares.LimitHandler(lmt))
	router.Use(middleware.Recoverer)

	//store, err := redis.New(opts.Redis.Addresses...)
	//if err != nil {
	//	log.Fatalf("redis.New(): %v", err)
	//}

	store, err := badger.New(opts.Badger.Path)
	if err != nil {
		log.Fatalf("badger.New(): %v", err)
	}

	apiSrv := handlers.New(store, opts.Domain, opts.Expiry, func(s string) bool {
		return true
	})

	router.Post("/", apiSrv.CreateShortLink)
	router.Get("/{id}", apiSrv.HandleRedirect)

	server := http.Server{Addr: opts.Host + ":" + opts.Port, Handler: router}

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
			_ = store.Shutdown()
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
	log.Infof("Listening to HTTP requests at %s", server.Addr)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
