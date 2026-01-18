package queuemonitor

import "context"

// InactiveState represents the state when queue is not active (DUW off hours)
type InactiveState struct {
	notifier    Notifier
	channelName string
}

func (s *InactiveState) Name() string     { return "Inactive" }
func (s *InactiveState) TicketsLeft() int { return 0 }

func (s *InactiveState) Handle(ctx context.Context, queue *Queue) (QueueState, error) {
	if !queue.Active {
		return s, nil // Stay inactive without notification
	}

	// Queue has become active
	if err := sendNotification(ctx, s.notifier, s.channelName, queue, false); err != nil {
		return s, err
	}

	if queue.Enabled {
		return &ActiveEnabledState{notifier: s.notifier, channelName: s.channelName, ticketsLeft: queue.TicketsLeft}, nil
	}
	return &ActiveDisabledState{notifier: s.notifier, channelName: s.channelName}, nil
}
