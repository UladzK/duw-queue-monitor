package queuemonitor

import (
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/queuemonitor/notifications"
)

// DefaultQueueMonitor is responsible for collecting queue status and sending notifications about changes in queue availability.
// Essentially, it is a state machine that checks the queue status periodically and notifies about changes.
// It uses a StatusCollector to get the queue status and a Notifier to send notifications.
type DefaultQueueMonitor struct {
	cfg                *Config
	log                *logger.Logger
	collector          *StatusCollector
	notifier           notifications.Notifier
	isStateInitialized bool
	state              *MonitorState
}

func NewQueueMonitor(cfg *Config, log *logger.Logger, collector *StatusCollector, notifier notifications.Notifier) *DefaultQueueMonitor {
	return &DefaultQueueMonitor{
		cfg:                cfg,
		log:                log,
		collector:          collector,
		notifier:           notifier,
		isStateInitialized: false,
		state:              &MonitorState{},
	}
}

func (h *DefaultQueueMonitor) Init(initState *MonitorState) {
	if initState == nil {
		panic("QueueMonitor.Init called with nil state. This should not happen")
	}

	h.state = initState
	h.isStateInitialized = true

	h.log.Info("QueueMonitor initialized with state:", "initState", initState)
}

func (h *DefaultQueueMonitor) GetState() *MonitorState {
	return h.state
}

func (h *DefaultQueueMonitor) CheckAndProcessStatus() error {
	newState, err := h.collector.GetQueueStatus()
	if err != nil {
		return fmt.Errorf("error getting queue status: %w", err)
	}

	if !newState.Active {
		h.log.Debug("Queue is not active, skipping notification", "newState", newState)
		h.updateState(newState)
		return nil
	}

	// Notify if state is not initialized, or state changed
	shouldNotifyStatusUpdate := !h.isStateInitialized || h.stateChanged(newState)

	if shouldNotifyStatusUpdate {
		if err := h.notifier.SendGeneralQueueStatusUpdateNotification(newState.Name, newState.Active, newState.Enabled,
			newState.TicketValue, newState.TicketsLeft); err != nil {
			return fmt.Errorf("error sending queue enabled notifiication: %w", err)
		}
	}

	h.updateState(newState)
	h.log.Debug("Latest state:", "latestState", h.state)

	return nil
}

func (h *DefaultQueueMonitor) stateChanged(newState *Queue) bool {
	// Notify if status changed, or tickets left changed (when enabled)
	if h.statusChanged(newState) || (newState.Enabled && h.state.TicketsLeft != newState.TicketsLeft) {
		h.log.Debug("Sending notification. Conditions met for notification.", "is not initialized", !h.isStateInitialized, "status changed", h.statusChanged(newState),
			"tickets left changed", h.state.TicketsLeft != newState.TicketsLeft)
		h.log.Debug("Current state and new state", "currentState", h.state, "newState", newState)

		return true
	}

	return false
}

func (h *DefaultQueueMonitor) statusChanged(newQueueStatus *Queue) bool {
	return h.state.QueueActive != newQueueStatus.Active || h.state.QueueEnabled != newQueueStatus.Enabled
}

func (h *DefaultQueueMonitor) updateState(newQueueStatus *Queue) {
	h.isStateInitialized = true
	h.state.LastTicketProcessed = newQueueStatus.TicketValue
	h.state.QueueEnabled = newQueueStatus.Enabled
	h.state.QueueActive = newQueueStatus.Active
	h.state.TicketsLeft = newQueueStatus.TicketsLeft
}
