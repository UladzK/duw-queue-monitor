package statuscollector

import (
	"context"
	"fmt"
	"time"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// Runner is responsible for the main loop of the status collector which periodically checks the queue status using the QueueMonitor.
type Runner struct {
	cfg       *Config
	log       *logger.Logger
	monitor   *QueueMonitor
	stateRepo *MonitorStateRepository
}

func NewRunner(cfg *Config, log *logger.Logger, monitor *QueueMonitor, stateRepo *MonitorStateRepository) *Runner {
	return &Runner{
		cfg:       cfg,
		log:       log,
		monitor:   monitor,
		stateRepo: stateRepo,
	}
}

func (h *Runner) Run(ctx context.Context, done chan<- bool) {
	h.log.Info("Initializing monitor state")
	h.initMonitorState(ctx)

	h.log.Info("Started monitor loop")
	for {
		select {
		case <-ctx.Done():
			h.log.Info("Received shutdown signal. Saving monitor state and stopping monitor loop")
			h.saveMonitorState(ctx)

			h.log.Info("Stopped monitor loop")
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

func (h *Runner) saveMonitorState(ctx context.Context) {

	latestState := h.monitor.GetState()
	if latestState == nil {
		h.log.Error("Failed to save monitor state", fmt.Errorf("monitor state is nil"))
		return
	}

	// todo: ideally, there should timeout for saving state, but it is not critical. Redis in-cluster is super reliable.
	saveCtx := context.WithoutCancel(ctx)

	if err := h.stateRepo.Save(saveCtx, latestState); err != nil {
		h.log.Error("Failed to save monitor state to Redis", err)
		return
	}

	h.log.Info("Monitor state saved successfully to Redis")
}

func (h *Runner) initMonitorState(ctx context.Context) {
	latestState, err := h.stateRepo.Get(ctx)
	if err != nil {
		h.log.Error("failed to get latest monitor state from Redis", err)
	}

	if latestState == nil {
		latestState = &MonitorState{
			QueueActive:         false,
			QueueEnabled:        false,
			LastTicketProcessed: "",
			TicketsLeft:         0,
		}

		h.log.Info("No previous monitor state found, initializing with default values")
	}

	h.monitor.Init(latestState)
}
