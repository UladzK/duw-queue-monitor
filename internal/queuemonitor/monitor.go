package queuemonitor

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// DefaultQueueMonitor is responsible for collecting queue status and sending notifications about changes in queue availability.
// Essentially, it's a state machine which tracks the current state of the DUW queue.
type DefaultQueueMonitor struct {
	cfg       *Config
	log       *logger.Logger
	collector *StatusCollector
	notifier  Notifier
	state     QueueState
	lastQueue *Queue
}

func NewQueueMonitor(cfg *Config, log *logger.Logger, collector *StatusCollector, notifier Notifier) *DefaultQueueMonitor {
	m := &DefaultQueueMonitor{
		cfg:       cfg,
		log:       log,
		collector: collector,
		notifier:  notifier,
	}
	m.state = &UninitializedState{notifier: notifier, channelName: cfg.BroadcastChannelName}
	return m
}

func (h *DefaultQueueMonitor) Init(initState *MonitorState) {
	if initState == nil {
		panic("QueueMonitor.Init called with nil state. This should not happen")
	}

	h.state = StateFromPersistence(initState, h.notifier, h.cfg.BroadcastChannelName)
	h.log.Info("QueueMonitor initialized with state:", "stateName", h.state.Name(), "initState", initState)
}

func (h *DefaultQueueMonitor) GetState() *MonitorState {
	return StateToPersistence(h.state, h.lastQueue)
}

func (h *DefaultQueueMonitor) CheckAndProcessStatus(ctx context.Context) error {
	queue, err := h.collector.GetQueueStatus(ctx)
	if err != nil {
		return fmt.Errorf("error getting queue status: %w", err)
	}

	prevStateName := h.state.Name()
	newState, err := h.state.Handle(ctx, queue)
	if err != nil {
		return err
	}

	if newState.Name() != prevStateName {
		h.log.Info("State transition", "from", prevStateName, "to", newState.Name())
	}

	h.state = newState
	h.lastQueue = queue
	h.log.Debug("Latest state:", "stateName", h.state.Name(), "ticketsLeft", h.state.TicketsLeft())

	return nil
}
