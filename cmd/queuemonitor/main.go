package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/notifications"
	"uladzk/duw_kolejka_checker/internal/queuemonitor"

	"github.com/caarlos0/env/v11"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log, err := buildLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	runner, err := buildRunner(log)
	if err != nil {
		panic("failed to initialize runner: " + err.Error())
	}

	log.Info("Starting queue monitor...")

	done := make(chan bool, 1)
	go runner.Run(ctx, done)

	log.Info("Queue monitor started. Waiting for shutdown signal...")
	<-ctx.Done()
	log.Info("Received shutdown signal, waiting for status collector to stop...")
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
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	httpClient := &http.Client{}

	opt, err := redis.ParseURL(cfg.QueueMonitor.RedisConString)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(opt)

	stateRepo := queuemonitor.NewMonitorStateRepository(redisClient, cfg.QueueMonitor.StateTtlSeconds)
	collector := queuemonitor.NewStatusCollector(&cfg.QueueMonitor, httpClient, log)
	notifier := buildNotifier(&cfg, log, httpClient)
	monitor := queuemonitor.NewQueueMonitor(&cfg, log, collector, notifier)
	weekdayMonitor := queuemonitor.NewWeekdayQueueMonitor(monitor, queuemonitor.NewSystemDateTimeProvider(), log)

	runner := queuemonitor.NewRunner(&cfg, log, weekdayMonitor, stateRepo)
	return runner, nil
}

func buildNotifier(cfg *queuemonitor.Config, log *logger.Logger, httpClient *http.Client) queuemonitor.Notifier {
	return notifications.NewTelegramNotifier(&cfg.NotificationTelegram, log, httpClient)
}
