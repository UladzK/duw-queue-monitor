package queuemonitor

import (
	"context"
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// DefaultQueueMonitor is responsible for collecting queue status and sending notifications about changes in queue availability.
// It uses the State Pattern to manage queue state transitions.
// It uses a StatusCollector to get the queue status and a Notifier to send notifications.
type DefaultQueueMonitor struct {
	cfg       *Config
	log       *logger.Logger
	collector *StatusCollector
	notifier  Notifier
	state     QueueState
	lastQueue *Queue
}

func NewQueueMonitor(cfg *Config, log *logger.Logger, collector *StatusCollector, notifier Notifier) *DefaultQueueMonitor {
	return &DefaultQueueMonitor{
		cfg:       cfg,
		log:       log,
		collector: collector,
		notifier:  notifier,
		state:     &UninitializedState{},
	}
}

func (h *DefaultQueueMonitor) Init(initState *MonitorState) {
	if initState == nil {
		panic("QueueMonitor.Init called with nil state. This should not happen")
	}

	h.state = StateFromPersistence(initState)
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
	newState, err := h.state.Handle(ctx, h, queue)
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

// notifyQueueStatus sends a notification about the queue status.
// This is called by state implementations when they need to notify.
func (h *DefaultQueueMonitor) notifyQueueStatus(ctx context.Context, queue *Queue) error {
	channelName := fmt.Sprintf("@%s", h.cfg.BroadcastChannelName)
	message := buildQueueAvailableMsg(queue.Name, queue.Enabled, queue.TicketValue, queue.TicketsLeft)
	if err := h.notifier.SendMessage(ctx, channelName, message); err != nil {
		return fmt.Errorf("error sending queue notification: %w", err)
	}
	return nil
}
