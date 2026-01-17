package queuemonitor

import "context"

// UninitializedState represents the initial state before first check.
// Always notifies if queue is active on first check.
type UninitializedState struct{}

func (s *UninitializedState) Name() string     { return "Uninitialized" }
func (s *UninitializedState) TicketsLeft() int { return 0 }

func (s *UninitializedState) Handle(ctx context.Context, m *DefaultQueueMonitor, queue *Queue) (QueueState, error) {
	if !queue.Active {
		return &InactiveState{}, nil
	}

	// Queue is active - always notify on first check (preserves existing behavior)
	if err := m.notifyQueueStatus(ctx, queue); err != nil {
		return s, err
	}

	if queue.Enabled {
		return &ActiveEnabledState{ticketsLeft: queue.TicketsLeft}, nil
	}
	return &ActiveDisabledState{}, nil
}
