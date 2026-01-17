package queuemonitor

import "context"

// ActiveDisabledState represents the state when queue is active but no tickets left.
type ActiveDisabledState struct {
	notifier    Notifier
	channelName string
}

func (s *ActiveDisabledState) Name() string     { return "ActiveDisabled" }
func (s *ActiveDisabledState) TicketsLeft() int { return 0 }

func (s *ActiveDisabledState) Handle(ctx context.Context, queue *Queue) (QueueState, error) {
	if !queue.Active {
		// Transition to inactive - no notification
		return &InactiveState{notifier: s.notifier, channelName: s.channelName}, nil
	}

	if queue.Enabled {
		if err := sendNotification(ctx, s.notifier, s.channelName, queue); err != nil {
			return s, err
		}
		return &ActiveEnabledState{notifier: s.notifier, channelName: s.channelName, ticketsLeft: queue.TicketsLeft}, nil
	}

	return s, nil
}
