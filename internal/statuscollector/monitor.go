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
	cfg       *Config
	log       *logger.Logger
	collector *StatusCollector
	notifier  notifications.Notifier
	state     QueueState
}

type QueueState struct {
	isStateInitialized  bool
	queueActive         bool
	queueEnabled        bool
	lastTicketProcessed string
	ticketsLeft         int
}

func NewQueueMonitor(cfg *Config, log *logger.Logger, collector *StatusCollector, notifier notifications.Notifier) *QueueMonitor {
	return &QueueMonitor{
		cfg:       cfg,
		log:       log,
		collector: collector,
		notifier:  notifier,
		state: QueueState{
			isStateInitialized: false,
		},
	}
}

func (h *QueueMonitor) CheckAndProcessStatus() error {

	newState, err := h.collector.GetQueueStatus()
	if err != nil {
		return fmt.Errorf("error getting queue status: %w", err)
	}

	// If the monitor has just started while queue has not been enabled yet, we don't want to send any notifications.
	if !h.state.isStateInitialized && !newState.Active {
		return nil
	}

	if !h.state.isStateInitialized || h.statusChanged(newState) {
		if err := h.pushQueueEnabledNotification(newState); err != nil {
			return err
		}

		h.updateState(newState)

		return nil
	}

	if !newState.Enabled {
		return nil
	}

	if h.state.ticketsLeft != newState.TicketsLeft {
		if err := h.pushQueueEnabledNotification(newState); err != nil {
			return err
		}

		h.updateState(newState)
	}

	return nil
}

func (h *QueueMonitor) statusChanged(newQueueStatus *Queue) bool {
	return h.state.queueActive != newQueueStatus.Active || h.state.queueEnabled != newQueueStatus.Enabled
}

func (h *QueueMonitor) pushQueueEnabledNotification(newQueueStatus *Queue) error {
	if err := h.notifier.SendGeneralQueueStatusUpdateNotification(newQueueStatus.Name, newQueueStatus.Enabled, newQueueStatus.TicketValue, newQueueStatus.TicketsLeft); err != nil {
		return fmt.Errorf("error sending queue enabled notifiication: %w", err)
	}

	return nil
}

func (h *QueueMonitor) updateState(newQueueStatus *Queue) {
	h.state.isStateInitialized = true
	h.state.lastTicketProcessed = newQueueStatus.TicketValue
	h.state.queueEnabled = newQueueStatus.Enabled
	h.state.queueActive = newQueueStatus.Active
}
