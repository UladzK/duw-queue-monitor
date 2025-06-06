package statuscollector

import (
	"context"
	"fmt"
	"time"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// Runner is responsible for the main loop of the status collector which periodically checks the queue status
type Runner struct {
	cfg     *Config
	log     *logger.Logger
	monitor *QueueMonitor
}

func NewRunner(cfg *Config, log *logger.Logger) *Runner {
	return &Runner{
		cfg:     cfg,
		log:     log,
		monitor: NewQueueMonitor(cfg, log),
	}
}

func (h *Runner) Run(ctx context.Context, done chan<- bool) {
	for {
		select {
		case <-ctx.Done():
			h.log.Info("Received shutdown signal, exiting...")
			done <- true
			return
		default:
			if err := h.monitor.CheckAndProcessStatus(); err != nil {
				h.log.Error("Error during collecting status and pushing notifications", err)
			}

			h.log.Debug(fmt.Sprintf("Status collection is completed. Checking again in %v seconds", h.cfg.StatusCheckInternalSeconds))
			time.Sleep(time.Duration(h.cfg.StatusCheckInternalSeconds) * time.Second) // TODO: will be sleeping even after SIGTERM. ticket is the better option?
		}
	}
}
