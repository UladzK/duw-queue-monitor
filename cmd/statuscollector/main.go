package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"uladzk/duw_kolejka_checker/internal/statuscollector"

	"github.com/caarlos0/env/v11"
)

func main() {
	// TODO: add graceful shutdown to finalize status processing and save the state before exiting

	var cfg statuscollector.Config

	err := env.Parse(&cfg)

	if err != nil {
		panic("Failed to get environment variables: " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	handler := statuscollector.NewHandler(&cfg)

	fmt.Println("Status collector started")
	go handler.Run(ctx, done)

	<-sigChan
	fmt.Println("Received shutdown signal, stopping status collector...")
	cancel()
	<-done

	fmt.Println("Status collector stopped")
}
