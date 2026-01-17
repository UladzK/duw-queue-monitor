package queuemonitor

import "context"

// ActiveEnabledState represents the state when queue is active and accepting bookings.
type ActiveEnabledState struct {
	ticketsLeft int
}

func (s *ActiveEnabledState) Name() string     { return "ActiveEnabled" }
func (s *ActiveEnabledState) TicketsLeft() int { return s.ticketsLeft }

func (s *ActiveEnabledState) Handle(ctx context.Context, m *DefaultQueueMonitor, queue *Queue) (QueueState, error) {
	if !queue.Active {
		// Transition to inactive - no notification (matches current behavior)
		return &InactiveState{}, nil
	}

	if !queue.Enabled {
		// Transition: Enabled -> Disabled (notify user - queue now unavailable)
		if err := m.notifyQueueStatus(ctx, queue); err != nil {
			return s, err
		}
		return &ActiveDisabledState{}, nil
	}

	// Still enabled - check if tickets changed
	if queue.TicketsLeft != s.ticketsLeft {
		if err := m.notifyQueueStatus(ctx, queue); err != nil {
			return s, err
		}
		return &ActiveEnabledState{ticketsLeft: queue.TicketsLeft}, nil
	}

	return s, nil // No change
}
