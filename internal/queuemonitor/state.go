package queuemonitor

import (
	"context"
	"fmt"
)

// QueueState represents a state in the queue monitor state machine.
// Each state is responsible for handling incoming queue status and determining if a transition to a new state should occur.
type QueueState interface {
	// Handle processes the new queue status and handles state transitions.
	Handle(ctx context.Context, queue *Queue) (QueueState, error)

	// Name returns the state name for logging and persistence.
	Name() string

	// TicketsLeft returns the last known tickets count (only relevant for ActiveEnabledState).
	TicketsLeft() int
}

// sendNotification sends a notification about the queue status during state transitions.
func sendNotification(ctx context.Context, notifier Notifier, channelName string, queue *Queue, isInactive bool) error {
	chatID := fmt.Sprintf("@%s", channelName)
	var message string
	if isInactive {
		message = buildQueueInactiveMsg(queue.Name)
	} else {
		message = buildQueueAvailableMsg(queue.Name, queue.Enabled, queue.TicketValue, queue.TicketsLeft)
	}
	if err := notifier.SendMessage(ctx, chatID, message); err != nil {
		return fmt.Errorf("error sending queue notification: %w", err)
	}
	return nil
}

// StateFromPersistence reconstructs a QueueState from persisted MonitorState.
func StateFromPersistence(ms *MonitorState, notifier Notifier, channelName string) QueueState {
	if ms == nil {
		return &UninitializedState{notifier: notifier, channelName: channelName}
	}

	if ms.StateName != "" {
		switch ms.StateName {
		case "Inactive":
			return &InactiveState{notifier: notifier, channelName: channelName}
		case "ActiveDisabled":
			return &ActiveDisabledState{notifier: notifier, channelName: channelName}
		case "ActiveEnabled":
			return &ActiveEnabledState{notifier: notifier, channelName: channelName, ticketsLeft: ms.TicketsLeft}
		case "Uninitialized":
			return &UninitializedState{notifier: notifier, channelName: channelName}
		}
	}

	// For backwards compatibility: derive state from boolean flags
	if !ms.QueueActive {
		return &InactiveState{notifier: notifier, channelName: channelName}
	}
	if ms.QueueEnabled {
		return &ActiveEnabledState{notifier: notifier, channelName: channelName, ticketsLeft: ms.TicketsLeft}
	}
	return &ActiveDisabledState{notifier: notifier, channelName: channelName}
}

// StateToPersistence converts a QueueState to MonitorState for persistence.
func StateToPersistence(state QueueState, queue *Queue) *MonitorState {
	ms := &MonitorState{
		StateName: state.Name(),
	}

	// for backward compatibility
	switch state.Name() {
	case "Inactive":
		ms.QueueActive = false
		ms.QueueEnabled = false
	case "ActiveDisabled":
		ms.QueueActive = true
		ms.QueueEnabled = false
	case "ActiveEnabled":
		ms.QueueActive = true
		ms.QueueEnabled = true
		ms.TicketsLeft = state.TicketsLeft()
	case "Uninitialized":
		ms.QueueActive = false
		ms.QueueEnabled = false
	}

	if queue != nil {
		ms.LastTicketProcessed = queue.TicketValue
		ms.TicketsLeft = queue.TicketsLeft
	}

	return ms
}
