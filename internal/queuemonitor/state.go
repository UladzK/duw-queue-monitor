package queuemonitor

import "context"

// QueueState represents a state in the queue monitor state machine.
// Each state is responsible for handling incoming queue status and
// determining if a transition to a new state should occur.
type QueueState interface {
	// Handle processes the new queue status and returns the next state.
	// It may trigger notifications via the monitor's notification methods.
	Handle(ctx context.Context, m *DefaultQueueMonitor, queue *Queue) (QueueState, error)

	// Name returns the state name for logging and persistence.
	Name() string

	// TicketsLeft returns the last known tickets count (only relevant for ActiveEnabledState).
	TicketsLeft() int
}

// StateFromPersistence reconstructs a QueueState from persisted MonitorState.
// Handles both new format (StateName) and legacy format (boolean flags).
func StateFromPersistence(ms *MonitorState) QueueState {
	if ms == nil {
		return &UninitializedState{}
	}

	// New format: use StateName directly
	if ms.StateName != "" {
		switch ms.StateName {
		case "Inactive":
			return &InactiveState{}
		case "ActiveDisabled":
			return &ActiveDisabledState{}
		case "ActiveEnabled":
			return &ActiveEnabledState{ticketsLeft: ms.TicketsLeft}
		case "Uninitialized":
			return &UninitializedState{}
		}
	}

	// For backwards compatibility: derive state from boolean flags
	if !ms.QueueActive {
		return &InactiveState{}
	}
	if ms.QueueEnabled {
		return &ActiveEnabledState{ticketsLeft: ms.TicketsLeft}
	}
	return &ActiveDisabledState{}
}

// StateToPersistence converts a QueueState to MonitorState for persistence.
// Maintains backward compatibility by populating both new and legacy fields.
func StateToPersistence(state QueueState, queue *Queue) *MonitorState {
	ms := &MonitorState{
		StateName: state.Name(),
	}

	// Populate legacy fields for backward compatibility
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

	// If queue data is available, use it for ticket info
	if queue != nil {
		ms.LastTicketProcessed = queue.TicketValue
		ms.TicketsLeft = queue.TicketsLeft
	}

	return ms
}
