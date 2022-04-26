package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	storage "go-scheduler-demo/localstorage"
	"go-scheduler-demo/server"
	"go-scheduler-demo/validators"
)

func main() {
	// allow syscall to terminate
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	repo := storage.New(nil)
	jsonValidator := validators.JSONValidator{}

	srv := server.New(server.Config{
		Addr:      ":8080",
		Repo:      repo,
		Validator: &jsonValidator,
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal/cancellation.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of
	// shutdown; provide manual option for local deploys
	stop()
	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// Pause 1 sec to allow server to finish the request it is currently
	// handling.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown: ", err)
	}

	log.Println("server exiting")
}
