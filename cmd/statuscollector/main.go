package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/statuscollector"
	"uladzk/duw_kolejka_checker/internal/statuscollector/notifications"

	"github.com/caarlos0/env/v11"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log, err := buildLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	runner, err := buildRunner(log)
	if err != nil {
		panic("failed to initialize runner: " + err.Error())
	}

	log.Info("Status collector started")
	go runner.Run(ctx, done)

	<-sigChan
	log.Info("Received shutdown signal, stopping status collector...")
	cancel()
	<-done

	log.Info("Status collector stopped")
}

func buildLogger() (*logger.Logger, error) {
	var cfg logger.Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return logger.NewLogger(&cfg), nil
}

func buildRunner(log *logger.Logger) (*statuscollector.Runner, error) {
	var cfg statuscollector.Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			// needed because otherwise the TLS connection is not established when calling from inside the container. silly workaround which just works
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	collector := statuscollector.NewStatusCollector(&cfg.StatusCollector, httpClient)
	notifier := notifications.NewPushOverNotifier(&cfg.NotificationPushOver, log, httpClient)
	monitor := statuscollector.NewQueueMonitor(&cfg, log, collector, notifier)

	runner := statuscollector.NewRunner(&cfg, log, monitor)
	return runner, nil
}
