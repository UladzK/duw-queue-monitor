package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/statuscollector"

	"github.com/caarlos0/env/v11"
)

func main() {
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

	var logCfg logger.Config
	err = env.Parse(&logCfg)
	if err != nil {
		panic("Failed to get logger configuration: " + err.Error())
	}
	logger := logger.NewLogger(&logCfg)
	handler := statuscollector.NewHandler(&cfg, logger)

	logger.Info("Status collector started")
	go handler.Run(ctx, done)

	<-sigChan
	logger.Info("Received shutdown signal, stopping status collector...")
	cancel()
	<-done

	logger.Info("Status collector stopped")
}
