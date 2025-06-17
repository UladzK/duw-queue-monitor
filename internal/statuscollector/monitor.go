package statuscollector

import (
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
	"uladzk/duw_kolejka_checker/internal/statuscollector/notifications"
)

// QueueMonitor is responsible for collecting queue status and sending notifications about changes in queue availability.
// Essentially, it is a state machine that checks the queue status periodically and notifies about changes.
// It uses a StatusCollector to get the queue status and a Notifier to send notifications.
type QueueMonitor struct {
	cfg                *Config
	log                *logger.Logger
	collector          *StatusCollector
	notifier           notifications.Notifier
	isStateInitialized bool
	state              *MonitorState
}

func NewQueueMonitor(cfg *Config, log *logger.Logger, collector *StatusCollector, notifier notifications.Notifier) *QueueMonitor {
	return &QueueMonitor{
		cfg:                cfg,
		log:                log,
		collector:          collector,
		notifier:           notifier,
		isStateInitialized: false,
		state:              &MonitorState{},
	}
}

func (h *QueueMonitor) Init(initState *MonitorState) {
	if initState == nil {
		panic("QueueMonitor.Init called with nil state. This should not happen")
	}

	h.state = initState
	h.isStateInitialized = true

	h.log.Info("QueueMonitor initialized with state:", "initState", initState)
}

func (h *QueueMonitor) GetState() *MonitorState {
	return h.state
}

func (h *QueueMonitor) CheckAndProcessStatus() error {
	newState, err := h.collector.GetQueueStatus()
	if err != nil {
		return fmt.Errorf("error getting queue status: %w", err)
	}

	// Skip notification if not initialized (starting new day run) and queue is not active yet
	if !h.isStateInitialized && !newState.Active {
		h.log.Debug("Queue is not active and state is not initialized, skipping notification")

		h.updateState(newState)
		return nil
	}

	// Notify if state is not initialized, or state changed
	shouldNotifyStatusUpdate := !h.isStateInitialized || h.stateChanged(newState)

	if shouldNotifyStatusUpdate {
		if err := h.pushGeneralQueueStatusUpdateNotification(newState); err != nil {
			return err
		}
	}

	h.updateState(newState)
	h.log.Debug("Queue status updated:", "newState", newState)

	return nil
}

func (h *QueueMonitor) stateChanged(newState *Queue) bool {
	// Notify if status changed, or tickets left changed (when enabled)
	if h.statusChanged(newState) || (newState.Enabled && h.state.TicketsLeft != newState.TicketsLeft) {

		h.log.Debug("Sending notification. Conditions met for notification.", "is initialized", h.isStateInitialized, "status changed", h.statusChanged(newState),
			"tickets left changed", h.state.TicketsLeft != newState.TicketsLeft)
		h.log.Debug("Current state and new state", "currentState", h.state, "newState", newState)

		return true
	}

	return false
}

func (h *QueueMonitor) statusChanged(newQueueStatus *Queue) bool {
	return h.state.QueueActive != newQueueStatus.Active || h.state.QueueEnabled != newQueueStatus.Enabled
}

func (h *QueueMonitor) pushGeneralQueueStatusUpdateNotification(newQueueStatus *Queue) error {
	if err := h.notifier.SendGeneralQueueStatusUpdateNotification(newQueueStatus.Name, newQueueStatus.Active, newQueueStatus.Enabled,
		newQueueStatus.TicketValue, newQueueStatus.TicketsLeft); err != nil {
		return fmt.Errorf("error sending queue enabled notifiication: %w", err)
	}

	return nil
}

func (h *QueueMonitor) updateState(newQueueStatus *Queue) {
	h.isStateInitialized = true
	h.state.LastTicketProcessed = newQueueStatus.TicketValue
	h.state.QueueEnabled = newQueueStatus.Enabled
	h.state.QueueActive = newQueueStatus.Active
	h.state.TicketsLeft = newQueueStatus.TicketsLeft
}
