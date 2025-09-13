package queuemonitor

import (
	"fmt"
	"uladzk/duw_kolejka_checker/internal/logger"
)

// Message constants for queue status notifications
const (
	msgQueueAvailableGeneral = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🎟️ Ostatni przywołany bilet: <b>%s</b>\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueAvailableShort   = "🔔 Kolejka <b>%s</b> jest teraz dostępna!\n🧾 Pozostało biletów: <b>%d</b>"
	msgQueueUnavailable      = "💤 Kolejka <b>%s</b> jest obecnie niedostępna."
	parseMode                = "HTML"
)

// buildQueueAvailableMsg creates a formatted message based on queue status
func buildQueueAvailableMsg(queueName string, queueEnabled bool, actualTicket string, numberOfTicketsLeft int) string {
	if !queueEnabled {
		return fmt.Sprintf(msgQueueUnavailable, queueName)
	}

	if actualTicket == "" {
		return fmt.Sprintf(msgQueueAvailableShort, queueName, numberOfTicketsLeft)
	}
	return fmt.Sprintf(msgQueueAvailableGeneral, queueName, actualTicket, numberOfTicketsLeft)
}

// DefaultQueueMonitor is responsible for collecting queue status and sending notifications about changes in queue availability.
// Essentially, it is a state machine that checks the queue status periodically and notifies about changes.
// It uses a StatusCollector to get the queue status and a Notifier to send notifications.
type DefaultQueueMonitor struct {
	cfg                *Config
	log                *logger.Logger
	collector          *StatusCollector
	notifier           Notifier
	isStateInitialized bool
	state              *MonitorState
}

func NewQueueMonitor(cfg *Config, log *logger.Logger, collector *StatusCollector, notifier Notifier) *DefaultQueueMonitor {
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
		channelName := fmt.Sprintf("@%s", h.cfg.BroadcastChannelName)
		message := buildQueueAvailableMsg(newState.Name, newState.Enabled, newState.TicketValue, newState.TicketsLeft)
		if err := h.notifier.SendMessage(channelName, message); err != nil {
			return fmt.Errorf("error sending queue enabled notification: %w", err)
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