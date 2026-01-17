package queuemonitor

import "context"

// InactiveState represents the state when queue is not active.
type InactiveState struct{}

func (s *InactiveState) Name() string     { return "Inactive" }
func (s *InactiveState) TicketsLeft() int { return 0 }

func (s *InactiveState) Handle(ctx context.Context, m *DefaultQueueMonitor, queue *Queue) (QueueState, error) {
	if !queue.Active {
		return s, nil // Stay inactive
	}

	// Transition: Inactive -> Active (notify user)
	if err := m.notifyQueueStatus(ctx, queue); err != nil {
		return s, err
	}

	if queue.Enabled {
		return &ActiveEnabledState{ticketsLeft: queue.TicketsLeft}, nil
	}
	return &ActiveDisabledState{}, nil
}
