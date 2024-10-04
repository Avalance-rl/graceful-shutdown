package main

import (
	"context"
	"errors"
	"examples/graceful-shutdown/closer"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := runServer(ctx); err != nil {
		log.Fatal(err)
	}
}

func runServer(ctx context.Context) error {

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	c := &closer.Closer{}

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})
	c.Add(srv.Shutdown)

	c.Add(func(ctx context.Context) error {
		time.Sleep(4 * time.Second)

		return nil
	})
	c.Add(func(ctx context.Context) error {
		time.Sleep(4 * time.Second)

		return nil
	})
	c.Add(func(ctx context.Context) error {
		time.Sleep(4 * time.Second)

		return nil
	})
	c.Add(func(ctx context.Context) error {
		time.Sleep(4 * time.Second)

		return nil
	})
	c.Add(func(ctx context.Context) error {
		time.Sleep(4 * time.Second)

		return nil
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	log.Printf("Listening on port %s", srv.Addr)
	<-ctx.Done()
	log.Println("Shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Close(shutdownCtx); err != nil {
		return fmt.Errorf("closer: %v", err)
	}

	return nil
}
