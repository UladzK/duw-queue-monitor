package queuemonitor

import "context"

// ActiveDisabledState represents the state when queue is active but no tickets left.
type ActiveDisabledState struct{}

func (s *ActiveDisabledState) Name() string     { return "ActiveDisabled" }
func (s *ActiveDisabledState) TicketsLeft() int { return 0 }

func (s *ActiveDisabledState) Handle(ctx context.Context, m *DefaultQueueMonitor, queue *Queue) (QueueState, error) {
	if !queue.Active {
		// Transition to inactive - no notification
		return &InactiveState{}, nil
	}

	if queue.Enabled {
		if err := m.notifyQueueStatus(ctx, queue); err != nil {
			return s, err
		}
		return &ActiveEnabledState{ticketsLeft: queue.TicketsLeft}, nil
	}

	return s, nil // Stay disabled
}
