package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/queuemonitor"
	"uladzk/duw_kolejka_checker/internal/queuemonitor/notifications"

	"github.com/caarlos0/env/v11"
	"github.com/redis/go-redis/v9"
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

	log.Info("Queue monitor started")
	go runner.Run(ctx, done)

	<-sigChan
	log.Info("Received shutdown signal, stopping status collector...")
	cancel()
	<-done

	log.Info("Queue monitor stopped")
}

func buildLogger() (*logger.Logger, error) {
	var cfg logger.Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return logger.NewLogger(&cfg), nil
}

func buildRunner(log *logger.Logger) (*queuemonitor.Runner, error) {
	var cfg queuemonitor.Config
	// TODO: implement Load() method in Configs
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			// needed because otherwise the TLS connection is not established when calling from inside the container. silly workaround which just works
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	opt, err := redis.ParseURL(cfg.QueueMonitor.RedisConString)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(opt)

	stateRepo := queuemonitor.NewMonitorStateRepository(redisClient, cfg.QueueMonitor.StateTtlSeconds)
	collector := queuemonitor.NewStatusCollector(&cfg.QueueMonitor, httpClient)
	notifier := buildNotifier(&cfg, log, httpClient)
	monitor := queuemonitor.NewQueueMonitor(&cfg, log, collector, notifier)

	runner := queuemonitor.NewRunner(&cfg, log, monitor, stateRepo)
	return runner, nil
}

func buildNotifier(cfg *queuemonitor.Config, log *logger.Logger, httpClient *http.Client) notifications.Notifier {
	if cfg.UseTelegramNotifications {
		return notifications.NewTelegramNotifier(&cfg.NotificationTelegram, log, httpClient)
	}

	return notifications.NewPushOverNotifier(&cfg.NotificationPushOver, log, httpClient)
}
